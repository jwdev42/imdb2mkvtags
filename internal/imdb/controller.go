package imdb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"io"
	"net/url"
	"strconv"
	"strings"
)

type options struct {
	UseJsonLD      bool
	UseFullCredits bool
	Languages      []string
}

type Controller struct {
	u       *url.URL
	o       *options
	titleID string
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

	return &Controller{
		u: u,
		o: &options{
			UseJsonLD:      false,
			UseFullCredits: false,
			Languages:      []string{global.DefaultLanguageIMDB},
		},
		titleID: path[2],
	}, nil
}

//Parses controller options. Reconfigures the controller after parsing was successful.
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
			switch arg[0] {
			case "jsonld":
				if err := parseBool(arg[1], &r.o.UseJsonLD); err != nil {
					return fmt.Errorf("Malformed argument value: %s", pair)
				}
			case "fullcredits":
				if err := parseBool(arg[1], &r.o.UseFullCredits); err != nil {
					return fmt.Errorf("Malformed argument value: %s", pair)
				}
			default:
				return fmt.Errorf("Unknown argument: %s", arg[0])
			}
		}
	}

	//Parse language option
	if flags.Lang != nil && *flags.Lang != "" {
		r.o.Languages = strings.Split(*flags.Lang, global.DelimControllerArgs)
	}

	return nil
}

func (r *Controller) Scrape() (*tags.Movie, error) {
	var tags *tags.Movie
	body := new(bytes.Buffer)
	if err := ihttp.GetBody(nil, r.u.String(), body, r.o.Languages...); err != nil {
		return nil, err
	}

	if r.o.UseJsonLD {
		json, err := ExtractMovieSchema(body)
		if err != nil {
			return nil, err
		}
		tags = json.Convert(r.o.Languages[0])
	} else {
		if t, err := r.scrapeTitlePage(body); err != nil {
			return nil, err
		} else {
			tags = t
		}
	}

	if r.o.UseFullCredits {
		if err := r.scrapeFullCredits(tags); err != nil {
			global.Log.Error(err)
		}
	}

	return tags, nil
}

func (r *Controller) scrapeFullCredits(movie *tags.Movie) error {
	fetchFullCredits := func(u string) (io.Reader, error) {
		u, err := TitleUrl2CreditsUrl(r.u.String())
		if err != nil {
			return nil, err
		}
		body := new(bytes.Buffer)
		if err := ihttp.GetBody(nil, u, body, r.o.Languages...); err != nil {
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

func (r *Controller) scrapeTitlePage(src io.Reader) (*tags.Movie, error) {
	title, err := NewTitle(r, src)
	if err != nil {
		return nil, err
	}

	movie := new(tags.Movie)

	movie.Imdb = tags.UniLingual(r.titleID)
	movie.SetFieldCallback("Actors", title.Actors)
	movie.SetFieldCallback("ContentRating", title.ContentRating)
	movie.SetFieldCallback("Directors", title.Directors)
	movie.SetFieldCallback("Genres", title.Genres)
	movie.SetFieldCallback("Keywords", title.Keywords)
	movie.SetFieldCallback("ReleaseDate", title.ReleaseDate)
	movie.SetFieldCallback("Synopses", title.Synopsis)
	movie.SetFieldCallback("Titles", title.Title)
	movie.SetFieldCallback("Writers", title.Writers)

	return movie, nil
}
