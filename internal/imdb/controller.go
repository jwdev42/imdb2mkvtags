package imdb

import (
	"bytes"
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

const DefaultLanguage = "en-US"

type options struct {
	UseJsonLD   bool
	UseFullCast bool
	Languages   []string
}

type Controller struct {
	u *url.URL
	o *options
}

func NewController(rawurl string) (*Controller, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &Controller{
		u: u,
		o: &options{
			UseJsonLD:   true,
			UseFullCast: false,
			Languages:   []string{DefaultLanguage},
		},
	}, nil
}

//Parses controller options. Reconfigures the controller after parsing was successful.
func (r *Controller) SetOptions(flags *cmdline.Flags) error {
	const delimArgs = ":" //delimiter to separate different arguments.
	const delimKV = "="   //delimiter to separate each argument from its value.

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
		pairs := strings.Split(*flags.Opts, delimArgs)
		for _, pair := range pairs {
			arg := strings.Split(pair, delimKV)
			if len(arg) != 2 {
				return fmt.Errorf("Malformed argument: %s", pair)
			}
			switch arg[0] {
			case "jsonld":
				if err := parseBool(arg[1], &r.o.UseJsonLD); err != nil {
					return fmt.Errorf("Malformed argument value: %s", pair)
				}
			case "fullcast":
				if err := parseBool(arg[1], &r.o.UseFullCast); err != nil {
					return fmt.Errorf("Malformed argument value: %s", pair)
				}
			}
		}
	}

	//Parse language option
	if flags.Lang != nil && *flags.Lang != "" {
		r.o.Languages = strings.Split(*flags.Lang, delimArgs)
	}

	return nil
}

func (r *Controller) Scrape() (*tags.Movie, error) {
	var tags *tags.Movie
	req, err := ihttp.NewBareReq("GET", r.u.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(r.o.Languages) > 0 {
		if err := ihttp.SetReqAccLang(req, r.o.Languages...); err != nil {
			return nil, err
		}
	}
	body := new(bytes.Buffer)
	if err := ihttp.Body(nil, req, body); err != nil {
		return nil, err
	}

	if r.o.UseJsonLD {
		json, err := ExtractMovieSchema(body)
		if err != nil {
			return nil, err
		}
		tags = json.Convert()
	} else {
		if t, err := r.scrapeTitlePage(body); err != nil {
			return nil, err
		} else {
			tags = t
		}
	}

	if r.o.UseFullCast {
		if err := r.scrapeFullCredits(tags); err != nil {
			global.Log.Error(err)
		}
	}

	return tags, nil
}

func (r *Controller) scrapeFullCredits(tags *tags.Movie) error {
	fetchFullCredits := func(u string) (io.Reader, error) {
		u, err := TitleUrl2CreditsUrl(r.u.String())
		if err != nil {
			return nil, err
		}
		req, err := ihttp.NewBareReq("GET", u, nil)
		if err != nil {
			return nil, err
		}
		if len(r.o.Languages) > 0 {
			if err := ihttp.SetReqAccLang(req, r.o.Languages...); err != nil {
				return nil, err
			}
		}
		body := new(bytes.Buffer)
		if err := ihttp.Body(nil, req, body); err != nil {
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

	if actors, err := credits.Cast(); err != nil {
		global.Log.Error(fmt.Errorf("Fullcredits: Could not process cast table: %s", err))
	} else {
		tags.Actors = actors
	}

	if directors := credits.NamesByID("director"); directors != nil {
		tags.Directors = directors
	} else {
		global.Log.Notice("Fullcredits: No directors found")
	}
	if producers := credits.NamesByID("producer"); producers != nil {
		tags.Producers = producers
	} else {
		global.Log.Notice("Fullcredits: No producers found")
	}
	if writers := credits.NamesByID("writer"); writers != nil {
		tags.Writers = writers
	} else {
		global.Log.Notice("Fullcredits: No writers found")
	}
	return nil
}

func (r *Controller) scrapeTitlePage(src io.Reader) (*tags.Movie, error) {
	title, err := NewTitle(r, src)
	if err != nil {
		return nil, err
	}
	movie := new(tags.Movie)

	if rating, err := title.ContentRating(); err != nil {
		global.Log.Error(fmt.Errorf("Title page: No content rating data found: %s", err))
	} else {
		movie.ContentRating = make([]tags.MultiLingual, 1)
		movie.ContentRating[0] = *rating
	}

	if genres, err := title.Genres(); err != nil {
		global.Log.Error(fmt.Errorf("Title page: No genres found: %s", err))
	} else {
		movie.Genres = genres
	}

	if release, err := title.ReleaseDate(); err != nil {
		global.Log.Error(fmt.Errorf("Title page: No release date found: %s", err))
	} else {
		movie.ReleaseDate = release
	}

	if tag, err := title.Synopsis(); err != nil {
		global.Log.Error(fmt.Errorf("Title page: No synopsis found: %s", err))
	} else {
		movie.Synopses = make([]tags.MultiLingual, 1)
		movie.Synopses[0] = *tag
	}

	if title, err := title.Title(); err != nil {
		global.Log.Error(fmt.Errorf("Title page: No movie title found: %s", err))
	} else {
		movie.Titles = make([]tags.MultiLingual, 1)
		movie.Titles[0] = *title
	}
	return movie, nil
}
