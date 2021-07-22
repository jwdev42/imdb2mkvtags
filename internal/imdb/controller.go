package imdb

import (
	"bytes"
	"errors"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"net/url"
)

type Options struct {
	UseJsonLD   bool
	UseFullCast bool
	Languages   []string
}

type Controller struct {
	u *url.URL
	o *Options
}

func NewController(rawurl string) (*Controller, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	return &Controller{
		u: u,
		o: &Options{
			UseJsonLD:   true,
			UseFullCast: false,
			Languages:   []string{"en-US"},
		},
	}, nil
}

//Sets controller options, argument must be a ptr to imdb.Options. Always returns nil
func (r *Controller) SetOptions(options interface{}) error {
	r.o = options.(*Options)
	return nil
}

func (r *Controller) Scrape() (*tags.Movie, error) {
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
		return json.Convert(), nil
	} else {
		panic("UseJsonLD must be true at the moment")
	}
	return nil, errors.New("You aren't supposed to be here")
}
