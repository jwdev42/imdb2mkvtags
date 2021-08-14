package controller

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/cmdline"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"net/url"
)

type Controller interface {
	Scrape() (*tags.Movie, error)
	SetOptions(options *cmdline.Flags) error
}

type EmptyUrlScheme string

func (r EmptyUrlScheme) Error() string {
	return string(r)
}

func Pick(rawurl string) (Controller, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if err := validateUrlScheme(u.Scheme); err != nil {
		if _, ok := err.(EmptyUrlScheme); ok {
			u.Scheme = "https"
			global.Log.Notice("Url scheme was empty, will try again using \"https\"")
			return Pick(u.String())
		} else {
			return nil, fmt.Errorf("Input url: %s", err)
		}
	}

	switch u.Host {
	case "imdb.com":
		fallthrough
	case "www.imdb.com":
		return imdb.NewController(u.String())
	case "epsteindidntkillhimself.com":
		global.Log.Die("Epstein didn’t kill himself")
	}
	return nil, errors.New("Input url: Host not supported")
}

func validateUrlScheme(scheme string) error {
	switch scheme {
	//begin whitelist
	case "http":
		fallthrough
	case "https":
		return nil
	//begin blacklist
	case "":
		return EmptyUrlScheme("Empty url scheme detected")
	default:
		return fmt.Errorf("Url scheme \"%s\" not supported", scheme)
	}
	panic("you're not supposed to be here")
}
