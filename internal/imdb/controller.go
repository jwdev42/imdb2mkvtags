//This file is part of imdb2mkvtags ©2022 Jörg Walter

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
	"strconv"
	"strings"
	"time"
)

// IMDB-specific options passed via parameter "opts"
type options struct {
	UseJsonLD      bool
	UseFullCredits bool
	UseKeywords    bool
	KeywordLimit   int
}

type Controller struct {
	u           *url.URL
	o           *options
	lang        []*lcconv.LngCntry
	defaultLang *lcconv.LngCntry
	titleID     string
}

func NewController(rawurl string) (*Controller, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	//validate url
	path := strings.Split(u.Path, "/")
	if len(path) < 3 || path[1] != "title" || !IsTitleID(path[2]) {
		return nil, errors.New("Unsupported or invalid URL")
	}

	//rebuild url path
	u.Path = fmt.Sprintf("/title/%s/", path[2])
	global.Log.Debug(fmt.Sprintf("IMDB controller: New controller with url \"%s\"", u.String()))

	defaultLang, err := lcconv.NewLngCntry("en-US")
	if err != nil {
		panic("invalid default language hardcoded into the program")
	}

	return &Controller{
		u:           u,
		o:           new(options),
		defaultLang: defaultLang,
		titleID:     path[2],
	}, nil
}

// Return IMDB's default language (english).
func (r *Controller) DefaultLang() *lcconv.LngCntry {
	return r.defaultLang
}

// Return the language chosen by the user.
func (r *Controller) PreferredLang() *lcconv.LngCntry {
	return r.lang[0]
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

	//Parse scraper-specific options
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

	//Parse language option
	if flags.Lang == nil {
		panic("Assertion failed: Command line parsing unit must set a default value for language")
	}
	r.lang = flags.Lang

	return nil
}

func (r *Controller) Scrape() (*tags.Movie, error) {
	//get title page
	body := new(bytes.Buffer)
	if err := ihttp.GetBody(nil, r.u.String(), body, r.lang...); err != nil {
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
	fetchFullCredits := func(u string) (io.Reader, error) {
		u, err := TitleUrl2CreditsUrl(r.u.String())
		if err != nil {
			return nil, err
		}
		body := new(bytes.Buffer)
		if err := ihttp.GetBody(nil, u, body, r.lang...); err != nil {
			return nil, err
		}
		return body, nil
	}

	body, err := fetchFullCredits(r.u.String())
	if err != nil {
		return fmt.Errorf("Fullcredits: Could not fetch page: %s", err)
	}

	credits, err := NewCredits(body)
	if err != nil {
		return fmt.Errorf("Fullcredits: Could not parse document: %s", err)
	}

	movie.SetFieldCallback("Actors", credits.Actors)
	movie.SetFieldCallback("Directors", credits.NamesByIDCallback("director"))
	movie.SetFieldCallback("Producers", credits.NamesByIDCallback("producer"))
	movie.SetFieldCallback("Writers", credits.NamesByIDCallback("writer"))
	return nil
}

func (r *Controller) scrapeKeywordPage(movie *tags.Movie) error {
	//Parse keyword page
	global.Log.Debug("Scraping keyword page")
	keywords, err := ParseKeywordPage(r.u.String()+"keywords", r.PreferredLang())
	if err != nil {
		return err
	}
	//Set the limit of exported keywords if a limit was given
	var limit int
	if r.o.KeywordLimit > 0 && r.o.KeywordLimit < len(keywords) {
		limit = r.o.KeywordLimit
	} else {
		limit = len(keywords)
	}
	//Copy the requested amount of keywords into tag objects
	keywordsTag := make([]tags.MultiLingual, limit)
	for i := 0; i < limit; i++ {
		keywordsTag[i].Text = keywords[i].Name
		keywordsTag[i].Lang = r.DefaultLang().ISO6391() //At the moment keywords are in english only, if IMDB changes that, it must also be changed here to PreferredLanguage().
	}
	global.Log.Debug(fmt.Sprintf("scrapeKeywordPage: Adding %d keywords", len(keywordsTag)))
	//Deploy the keyword tags to the movie object
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
