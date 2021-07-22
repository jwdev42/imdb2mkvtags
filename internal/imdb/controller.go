package imdb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"net/url"
	"strconv"
	"strings"
)

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
			Languages:   []string{"en-US"},
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

	//Parse imdb options
	if flags.Imdb != nil && *flags.Imdb != "" {
		pairs := strings.Split(*flags.Imdb, delimArgs)
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
