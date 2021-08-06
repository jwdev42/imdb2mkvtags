//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb/schema"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
)

//represents title pages https://www.imdb.com/title/$titleID/
type Title struct {
	c    *Controller
	root *html.Node
}

func NewTitle(c *Controller, r io.Reader) (*Title, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Title{
		c:    c,
		root: root,
	}, nil
}

func (r *Title) textByTestID(id string) (string, error) {
	const key = "data-testid"
	node := rottensoup.FirstElementByAttr(r.root, html.Attribute{Key: key, Val: id})
	if node == nil {
		return "", fmt.Errorf("No element found with attribute %s=\"%s\"", key, id)
	}
	text := rottensoup.FirstNodeByType(node, html.TextNode)
	if text == nil {
		return "", fmt.Errorf("No text node found that is a child of element with attribute %s=\"%s\"", key, id)
	}
	return text.Data, nil
}

func (r *Title) Synopsis() (*tags.MultiLingual, error) {
	const entity = "synopsis"
	text, err := r.textByTestID("plot-xl")
	if err != nil {
		return nil, fmt.Errorf("Could not fetch %s: %s", entity, err)
	}
	return &tags.MultiLingual{
		Text: text,
		Lang: "en",
	}, nil
}

func (r *Title) Title() (*tags.MultiLingual, error) {
	const entity = "title"
	text, err := r.textByTestID("hero-title-block__title")
	if err != nil {
		return nil, fmt.Errorf("Could not fetch %s: %s", entity, err)
	}
	return &tags.MultiLingual{
		Text: text,
		Lang: r.c.o.Languages[0],
	}, nil
}

func ScrapeTitlePage(c *Controller, src io.Reader) (*tags.Movie, error) {
	title, err := NewTitle(c, src)
	if err != nil {
		return nil, err
	}
	movie := new(tags.Movie)
	if tag, err := title.Synopsis(); err != nil {
		global.Log.Error(fmt.Errorf("Failed receiving synopsis: %s", err))
	} else {
		movie.Synopses = make([]tags.MultiLingual, 1)
		movie.Synopses[0] = *tag
	}

	if title, err := title.Title(); err != nil {
		global.Log.Error(fmt.Errorf("Failed receiving title: %s", err))
	} else {
		movie.Titles = make([]tags.MultiLingual, 1)
		movie.Titles[0] = *title
	}
	return movie, nil
}

//Scrapes the json-ld data from an imdb page and loads it into a movie schema object.
func ExtractMovieSchema(src io.Reader) (*schema.Movie, error) {
	root, err := html.Parse(src)
	if err != nil {
		return nil, err
	}

	head := rottensoup.FirstElementByTag(root, atom.Head)
	if head == nil {
		return nil, errors.New("No html head tag found")
	}

	schemas := rottensoup.ElementsByTagAndAttr(root, atom.Script, html.Attribute{Key: "type", Val: "application/ld+json"})
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
