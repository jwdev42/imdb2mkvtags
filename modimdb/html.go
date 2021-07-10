//This file is part of imdb2mkvtags ©2021 Jörg Walter

package modimdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/modimdb/schema"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
)

func ExtractMovieSchema(src io.Reader) (*schema.Movie, error) {
	root, err := html.Parse(src)
	if err != nil {
		return nil, err
	}

	head := rottensoup.FirstElementByTag(root, atom.Head)
	if head == nil {
		return nil, errors.New("No html head tag found")
	}

	schemas := rottensoup.ElementsByTagAndAttr(root, "script", html.Attribute{Key: "type", Val: "application/ld+json"})
	if len(schemas) < 1 {
		return nil, errors.New("No movie schema found")
	}
	jsonText := schemas[0].FirstChild.Data
	movie := new(schema.Movie)
	if err := json.Unmarshal([]byte(jsonText), &movie); err != nil {
		return nil, fmt.Errorf("Json unmarshaler: %s", err)
	}
	return movie, nil
}
