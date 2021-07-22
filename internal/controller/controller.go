package controller

import (
	"errors"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"net/url"
)

type Controller interface {
	Scrape() (*tags.Movie, error)
	SetOptions(options interface{}) error
}

func Pick(rawurl string) (Controller, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, errors.New("Unsupported url scheme")
	}

	switch u.Host {
	case "imdb.com":
		fallthrough
	case "www.imdb.com":
		return imdb.NewController(rawurl)
	}
	return nil, errors.New("url not supported")
}
