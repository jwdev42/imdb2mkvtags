//This file is part of imdb2mkvtags ©2022 - 2026 Jörg Walter

package imdb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/lcconv"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"io"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// Holds IMDB-specific options passed via parameter "opts".
// Also holds common opts that need to be known.
type options struct {
	UseJsonLD      bool
	UseFullCredits bool
	UseKeywords    bool
	KeywordLimit   int
	UserAgent      string // User Agent for HTTP client
}

type Controller struct {
	urlScheme   string
	urlCountry  string
	o           *options
	lang        []*lcconv.LngCntry
	defaultLang *lcconv.LngCntry
	titleID     string
}

func NewController(rawurl string) (*Controller, error) {
	const urlValidationFailed = "IMDB URL validation failed"
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// Set american english as default language
	defaultLang, err := lcconv.NewLngCntry("en-US")
	if err != nil {
		panic("invalid default language hardcoded into the program")
	}
	// Create controller
	cntrl := &Controller{
		urlScheme:   u.Scheme,
		o:           new(options),
		lang:        make([]*lcconv.LngCntry, 0),
		defaultLang: defaultLang,
	}
	// Validate url
	if u.Scheme == "imdb" {
		// Handling special scheme "imdb://"
		if !IsTitleID(u.Host) {
			global.Log.Errorf("The URL's host name must be a valid IMDB movie ID when scheme %q is in use", u.Scheme)
			return nil, errors.New(urlValidationFailed)
		}
		cntrl.urlScheme = "https"
		cntrl.titleID = u.Host
		return cntrl, nil
	}
	// Handling normal URLs
	if !path.IsAbs(u.Path) {
		return nil, errors.New("IMDB URL must have an absolute path")
	}
	path := strings.Split(u.Path, "/")
	if len(path) >= 4 && path[2] == "title" && IsTitleID(path[3]) {
		cntrl.titleID = path[3]
	} else if len(path) >= 3 && path[1] == "title" && IsTitleID(path[2]) {
		cntrl.titleID = path[2]
	} else {
		return nil, errors.New(urlValidationFailed)
	}
	return cntrl, nil
}

// Return IMDB's default language.
// At the moment this is hard-coded to english.
func (r *Controller) DefaultLang() *lcconv.LngCntry {
	return r.defaultLang
}

// Return the most significant language chosen by the user if one was set.
// Return the default language otherwise.
func (r *Controller) PreferredLang() *lcconv.LngCntry {
	if len(r.lang) == 0 {
		return r.defaultLang
	}
	return r.lang[0]
}

// Return the controller's title URL.
func (r *Controller) TitleURL() string {
	if r.urlCountry != "" {
		return fmt.Sprintf("%s://imdb.com/%s/title/%s",
			r.urlScheme, url.PathEscape(r.urlCountry), url.PathEscape(r.titleID))
	}
	return fmt.Sprintf("%s://imdb.com/title/%s", r.urlScheme, url.PathEscape(r.titleID))
}

// Return the controller's credits page URL.
func (r *Controller) CreditsURL() string {
	return r.TitleURL() + "/fullcredits"
}

// Return the controller's keywords page URL.
func (r *Controller) KeywordsURL() string {
	return r.TitleURL() + "/keywords"
}

// Parses controller options. Reconfigures the controller after parsing was successful.
func (r *Controller) SetOptions(flags *cmdline.Flags) error {

	parseBool := func(str string, val *bool) error {
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		*val = b
		return nil
	}

	// Set user agent
	r.o.UserAgent = *flags.UserAgent

	// Parse scraper-specific options
	if flags.Opts != nil && *flags.Opts != "" {
		pairs := strings.Split(*flags.Opts, global.DelimControllerArgs)
		for _, pair := range pairs {
			arg := strings.Split(pair, global.DelimControllerKV)
			if len(arg) != 2 {
				return fmt.Errorf("Malformed argument: %s", pair)
			}
			const malformedVal = "Malformed argument value: %s"
			switch arg[0] {
			case "jsonld":
				if err := parseBool(arg[1], &r.o.UseJsonLD); err != nil {
					return fmt.Errorf(malformedVal, pair)
				}
			case "fullcredits":
				if err := parseBool(arg[1], &r.o.UseFullCredits); err != nil {
					return fmt.Errorf(malformedVal, pair)
				}
			case "keywords":
				if err := parseBool(arg[1], &r.o.UseKeywords); err != nil {
					return fmt.Errorf(malformedVal, pair)
				}
			case "keyword-limit":
				limit, err := strconv.Atoi(arg[1])
				if err != nil {
					return fmt.Errorf("Illegal argument for %s", arg[0])
				}
				r.o.KeywordLimit = limit
			default:
				return fmt.Errorf("Unknown argument: %s", arg[0])
			}
		}
	}

	// Parse language option
	if flags.Lang != nil {
		r.lang = flags.Lang
	}

	return nil
}

func (r *Controller) Scrape() (*tags.Movie, error) {
	// get title page
	body := new(bytes.Buffer)
	if err := ihttp.GetBody(nil, r.o.UserAgent, r.TitleURL(), body, r.lang...); err != nil {
		return nil, err
	}

	var movie *tags.Movie

	if r.o.UseJsonLD {
		json, err := ExtractMovieSchema(body)
		if err != nil {
			return nil, err
		}
		movie = json.Convert(r.PreferredLang(), r.DefaultLang())
	} else {
		if t, err := r.scrapeTitlePage(body); err != nil {
			return nil, err
		} else {
			movie = t
		}
	}

	if r.o.UseFullCredits {
		if err := r.scrapeFullCredits(movie); err != nil {
			global.Log.Error(fmt.Errorf("Could not scrape full credits: %s", err))
		}
	}

	if r.o.UseKeywords {
		if err := r.scrapeKeywordPage(movie); err != nil {
			global.Log.Error(fmt.Errorf("Could not scrape keywords: %s", err))
		}
	}

	movie.Imdb = tags.UniLingual(r.titleID)
	movie.DateTagged = tags.UniLingual(time.Now().Format("2006-01-02"))

	return movie, nil
}

func (r *Controller) scrapeFullCredits(movie *tags.Movie) error {
	global.Log.Debug("Scraping credits page")
	fetchFullCredits := func() (io.Reader, error) {
		body := new(bytes.Buffer)
		if err := ihttp.GetBody(nil, r.o.UserAgent, r.CreditsURL(), body, r.lang...); err != nil {
			return nil, err
		}
		return body, nil
	}

	body, err := fetchFullCredits()
	if err != nil {
		return fmt.Errorf("Fullcredits: Could not fetch page: %s", err)
	}

	credits, err := NewCredits(body)
	if err != nil {
		return fmt.Errorf("Fullcredits: Could not parse document: %s", err)
	}

	movie.SetFieldCallback("Actors", credits.Actors)
	movie.SetFieldCallback("Directors", credits.Directors)
	movie.SetFieldCallback("Producers", credits.Producers)
	movie.SetFieldCallback("Writers", credits.Writers)
	return nil
}

func (r *Controller) scrapeKeywordPage(movie *tags.Movie) error {
	// Parse keyword page
	global.Log.Debug("Scraping keyword page")
	keywords, err := ParseKeywordPage(r.o.UserAgent, r.KeywordsURL(), r.PreferredLang())
	if err != nil {
		return err
	}
	// Set the limit of exported keywords if a limit was given
	var limit int
	if r.o.KeywordLimit > 0 && r.o.KeywordLimit < len(keywords) {
		limit = r.o.KeywordLimit
	} else {
		limit = len(keywords)
	}
	// Copy the requested amount of keywords into tag objects
	keywordsTag := make([]tags.MultiLingual, limit)
	for i := 0; i < limit; i++ {
		keywordsTag[i].Text = keywords[i].Name
		keywordsTag[i].Lang = r.DefaultLang().ISO6391() // At the moment keywords are in english only, if IMDB changes that, it must also be changed here to PreferredLanguage().
	}
	global.Log.Debug(fmt.Sprintf("scrapeKeywordPage: Adding %d keywords", len(keywordsTag)))
	// Deploy the keyword tags to the movie object
	movie.Keywords = keywordsTag
	return nil
}

func (r *Controller) scrapeTitlePage(src io.Reader) (*tags.Movie, error) {
	title, err := NewTitle(r, src)
	if err != nil {
		return nil, err
	}

	movie := new(tags.Movie)

	movie.SetFieldCallback("Actors", title.Actors)
	movie.SetFieldCallback("DateReleased", title.DateReleased)
	movie.SetFieldCallback("Directors", title.Directors)
	movie.SetFieldCallback("Genres", title.Genres)
	movie.SetFieldCallback("Synopses", title.Synopsis)
	movie.SetFieldCallback("Titles", title.Title)
	movie.SetFieldCallback("Writers", title.Writers)

	country := &tags.Country{Name: r.PreferredLang().Alpha3()}
	country.SetFieldCallback("LawRating", title.LawRating)

	if !country.IsEmpty() {
		movie.Countries = []*tags.Country{country}
	}

	return movie, nil
}
