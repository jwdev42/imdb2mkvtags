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

const DefaultLanguage = "en-US"

type options struct {
	UseJsonLD   bool
	UseFullCast bool
	Languages   []string
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
			UseJsonLD:   false,
			UseFullCast: false,
			Languages:   []string{DefaultLanguage},
		},
		titleID: path[2],
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

	const errNotFound = "Title page: No %s found"

	movie := new(tags.Movie)

	movie.Imdb = tags.UniLingual(r.titleID)

	setML := func(name, dataDesc string, f func() (*tags.MultiLingual, error)) {
		if v, err := f(); err != nil {
			global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, dataDesc), err))
		} else {
			movie.SetField(name, []tags.MultiLingual{*v})
		}
	}

	if actors, err := title.Actors(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "actors"), err))
	} else {
		movie.Actors = actors
	}

	setML("ContentRating", "content rating data", title.ContentRating)

	if directors, err := title.Directors(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "directors"), err))
	} else {
		movie.Directors = directors
	}

	if genres, err := title.Genres(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "genre information"), err))
	} else {
		movie.Genres = genres
	}

	if keywords, err := title.Keywords(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "keywords"), err))
	} else {
		movie.Keywords = keywords
	}

	if release, err := title.ReleaseDate(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "release date"), err))
	} else {
		movie.ReleaseDate = release
	}

	setML("Synopses", "synopsis", title.Synopsis)
	setML("Titles", "movie title", title.Title)

	if writers, err := title.Writers(); err != nil {
		global.Log.Error(fmt.Errorf("%s: %s", fmt.Sprintf(errNotFound, "writers"), err))
	} else {
		movie.Writers = writers
	}

	return movie, nil
}
