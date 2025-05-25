package controller

import (
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
			global.Log.Noticef("No URL scheme provided, using default %q", u.Scheme)
			return Pick(u.String())
		} else {
			return nil, fmt.Errorf("Input url: %s", err)
		}
	}
	// Pick by scheme first
	switch u.Scheme {
	case "imdb":
		return imdb.NewController(u.String())
	}
	// Pick by host when no special scheme was found
	switch u.Host {
	case "imdb.com", "www.imdb.com":
		return imdb.NewController(u.String())
	case "themoviedb.org", "www.themoviedb.org":
		global.Log.Die("TMDB is not supported yet")
	case "epsteindidntkillhimself.com":
		global.Log.Die("Epstein didn’t kill himself")
	}
	return nil, fmt.Errorf("Scraping host %q is not supported")
}

func validateUrlScheme(scheme string) error {
	switch scheme {
	// allowed
	case "http", "https", "imdb":
		return nil
	// denied
	case "":
		return EmptyUrlScheme("Empty url scheme detected")
	default:
		return fmt.Errorf("Url scheme \"%s\" not supported", scheme)
	}
	panic("you're not supposed to be here")
}
