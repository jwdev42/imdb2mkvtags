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

const attrTestID = "data-testid"

//represents title pages https://www.imdb.com/title/$titleID/
type Title struct {
	c       *Controller
	root    *html.Node
	credits creditsList
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

func (r *Title) parseCreditsList() error {
	list, err := parseCreditsList(r.root)
	if err != nil {
		return err
	}
	r.credits = list
	return nil
}

//Scrapes actors and their characters from the title page.
func (r *Title) Actors() ([]tags.Actor, error) {
	//Closure for scraping the actor's character
	scrapeCharacter := func(node *html.Node) (string, error) {
		const testIDCharacter = "cast-item-characters-link"
		entry := rottensoup.FirstElementByAttr(node, html.Attribute{Key: "data-testid", Val: testIDCharacter})
		if entry == nil {
			return "", errors.New("No character entry available")
		}
		textNode := rottensoup.FirstNodeByType(entry, html.TextNode)
		if textNode == nil {
			return "", errors.New("Character entry contains no text")
		}
		return textNode.Data, nil
	}
	const testIDActors = "title-cast-item__actor"
	entries := rottensoup.ElementsByAttr(r.root, html.Attribute{Key: "data-testid", Val: testIDActors})
	if entries == nil {
		return nil, fmt.Errorf("No actors found with testid '%s'", testIDActors)
	}
	actors := make([]tags.Actor, 0, len(entries))
	for _, entry := range entries {
		actor := tags.Actor{}
		//Find the text node with the actor's name
		textNode := rottensoup.FirstNodeByType(entry, html.TextNode)
		if textNode == nil {
			global.Log.Info("Skipping Actor entry without text node")
			continue
		}
		actor.Name = textNode.Data
		//Look up the character the actor plays
		character, err := scrapeCharacter(entry.Parent)
		if err != nil {
			global.Log.Info(fmt.Errorf("Could not find the character played by %s: %s", actor.Name, err))
		} else {
			actor.Character = character
		}
		//Adding actor to the list of actors
		actors = append(actors, actor)
	}
	if len(actors) < 1 {
		return nil, errors.New("No actors found: No text node in at least one actor data field")
	}
	return actors, nil
}

func (r *Title) LawRating() (tags.UniLingual, error) {
	text, err := r.extractFromHeroTitleBlock(1)
	if err != nil {
		return "", err
	}
	return tags.UniLingual(text), nil
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
			genres = append(genres, tags.MultiLingual{Text: text.Data, Lang: r.c.PreferredLang().ISO6391()})
		}
	}
	if len(genres) < 1 {
		return nil, errors.New(errNoGenreData)
	}
	return genres, nil
}

func (r *Title) Directors() ([]tags.UniLingual, error) {
	if r.credits == nil {
		if err := r.parseCreditsList(); err != nil {
			return nil, err
		}
	}
	names := r.credits["directors"]
	if len(names) < 1 {
		return nil, errors.New("No data available")
	}
	directors := make([]tags.UniLingual, len(names))
	for i, name := range names {
		directors[i] = tags.UniLingual(name)
	}
	return directors, nil
}

func (r *Title) DateReleased() (tags.UniLingual, error) {
	text, err := r.extractFromHeroTitleBlock(0)
	if err != nil {
		return "", err
	}
	return tags.UniLingual(text), nil
}

func (r *Title) Synopsis() ([]tags.MultiLingual, error) {
	val, err := r.testID2MultiLingual("plot-xl", r.c.PreferredLang().ISO6391())
	if err != nil {
		return nil, err
	}
	return []tags.MultiLingual{*val}, err
}

func (r *Title) Title() ([]tags.MultiLingual, error) {
	val, err := r.testID2MultiLingual("hero__pageTitle", r.c.PreferredLang().ISO6391())
	if err != nil {
		return nil, err
	}
	return []tags.MultiLingual{*val}, err
}

func (r *Title) Writers() ([]tags.UniLingual, error) {
	if r.credits == nil {
		if err := r.parseCreditsList(); err != nil {
			return nil, err
		}
	}
	names := r.credits["writers"]
	if len(names) < 1 {
		return nil, errors.New("No data available")
	}
	writers := make([]tags.UniLingual, len(names))
	for i, name := range names {
		writers[i] = tags.UniLingual(name)
	}
	return writers, nil
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

func (r *Title) extractFromHeroTitleBlock(num int) (string, error) {
	title, err := r.elementByTestID("hero__pageTitle")
	if err != nil {
		return "", err
	}
	ul := rottensoup.FirstElementByTag(title.Parent, atom.Ul)
	if ul == nil {
		return "", errors.New("")
	}
	items := rottensoup.ElementsByTag(ul, atom.Li)
	if items == nil {
		return "", errors.New("Hero-Title-Block list contains no elements")
	}
	if len(items) < num {
		return "", fmt.Errorf("Hero-Title-Block list index out of bounds (max: %d, is: %d)", len(items), num)
	}
	spans := rottensoup.ElementsByTag(items[num], atom.A)
	if spans == nil {
		return "", fmt.Errorf("No link element found in Hero-Title-Block list item %d", num)
	}
	text := rottensoup.FirstNodeByType(spans[0], html.TextNode)
	if text == nil {
		return "", errors.New("No text node found")
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
