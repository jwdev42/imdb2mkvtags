//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/imdb/schema"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
)

const attrTestID = "data-testid"

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

func (r *Title) Genres() ([]tags.MultiLingual, error) {
	const errNoGenreData = "No genre data available"
	node, err := r.elementByTestID("genres")
	if err != nil {
		return nil, err
	}
	spans := rottensoup.ElementsByTagAndAttr(node, atom.Span, html.Attribute{Key: "class", Val: "ipc-chip__text"})
	if len(spans) < 1 {
		return nil, errors.New(errNoGenreData)
	}
	genres := make([]tags.MultiLingual, 0, len(spans))
	for _, span := range spans {
		if text := rottensoup.FirstNodeByType(span, html.TextNode); text != nil {
			genres = append(genres, tags.MultiLingual{Text: text.Data, Lang: DefaultLanguage})
		}
	}
	if len(genres) < 1 {
		return nil, errors.New(errNoGenreData)
	}
	return genres, nil
}

func (r *Title) ReleaseDate() (tags.UniLingual, error) {
	e := rottensoup.ElementsByTagAndAttr(r.root, atom.Span, html.Attribute{Key: "class", Val: "TitleBlockMetaData__ListItemText-sc-12ein40-2 jedhex"})
	if e == nil {
		return "", errors.New("The html element that contains the release date was not found")
	}
	text := rottensoup.FirstNodeByType(e[0], html.TextNode)
	if text == nil {
		return "", errors.New("No text node found")
	}
	return tags.UniLingual(text.Data), nil
}

func (r *Title) Synopsis() (*tags.MultiLingual, error) {
	return r.testID2MultiLingual("plot-xl", DefaultLanguage)
}

func (r *Title) Title() (*tags.MultiLingual, error) {
	return r.testID2MultiLingual("hero-title-block__title", r.c.o.Languages[0])
}

func (r *Title) testID2MultiLingual(testID, lang string) (*tags.MultiLingual, error) {
	text, err := r.textByTestID(testID)
	if err != nil {
		return nil, err
	}
	return &tags.MultiLingual{
		Text: text,
		Lang: lang,
	}, nil
}

func (r *Title) elementByTestID(id string) (*html.Node, error) {
	node := rottensoup.FirstElementByAttr(r.root, html.Attribute{Key: attrTestID, Val: id})
	if node == nil {
		return nil, fmt.Errorf("No element found with attribute %s=\"%s\"", attrTestID, id)
	}
	return node, nil
}

func (r *Title) textByTestID(id string) (string, error) {
	node, err := r.elementByTestID(id)
	if err != nil {
		return "", err
	}
	text := rottensoup.FirstNodeByType(node, html.TextNode)
	if text == nil {
		return "", fmt.Errorf("No text node found that is a child of element with attribute %s=\"%s\"", attrTestID, id)
	}
	return text.Data, nil
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
