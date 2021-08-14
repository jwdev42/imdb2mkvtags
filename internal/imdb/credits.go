package imdb

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"os"
	"strings"
)

//represents "fullcredits" pages https://www.imdb.com/title/$titleID/fullcredits
type Credits struct {
	root *html.Node
}

func NewCredits(r io.Reader) (*Credits, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Credits{
		root: root,
	}, nil
}

func (r *Credits) Actors() ([]tags.Actor, error) {

	//Get the cast table
	var table *html.Node
	{
		tables := rottensoup.ElementsByTagAndAttr(r.root, atom.Table, html.Attribute{Key: "class", Val: "cast_list"})
		if len(tables) < 1 {
			return nil, errors.New("No cast table found")
		}
		table = tables[0]
	}

	actors := make([]tags.Actor, 0, 10)
	rows := rottensoup.ElementsByTagAndAttr(table, atom.Td, html.Attribute{Key: "class", Val: "primary_photo"})
	for i, row := range rows {
		actor, err := r.actor(row)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cast table row %d: %s\n", i+1, err)
			continue
		}
		actors = append(actors, *actor)
	}
	if len(actors) < 1 {
		return nil, errors.New("No actors found in cast table")
	}
	return actors, nil
}

func (r *Credits) NamesByIDCallback(id string) func() ([]tags.UniLingual, error) {
	return func() ([]tags.UniLingual, error) {
		nodes := r.namesByID(id)
		if nodes == nil {
			return nil, fmt.Errorf("Found no heading with ID \"%s\"", id)
		}
		names := make([]tags.UniLingual, 0, len(nodes))
		for i, n := range nodes {
			text, err := r.textFromNameCell(n)
			if err != nil {
				global.Log.Error(fmt.Errorf("Row %d of table \"%s\": %s", i+1, id, err))
			} else {
				names = append(names, tags.UniLingual(text))
			}
		}
		if len(names) > 0 {
			return names, nil
		}
		return nil, errors.New("No data found")
	}
}

func (r *Credits) actor(firstCol *html.Node) (*tags.Actor, error) {

	trimmedCellText := func(cell *html.Node) (string, error) {
		link := rottensoup.FirstElementByTag(cell, atom.A)
		if link == nil {
			return "", errors.New("No hyperlink found inside table cell")
		}
		text := link.FirstChild
		if text == nil || text.Type != html.TextNode {
			return "", errors.New("Hyperlink contains no text")
		}
		trimmed := strings.TrimSpace(text.Data)
		if len(trimmed) < 1 {
			return "", errors.New("Hyperlink text is empty")
		}
		return trimmed, nil
	}

	//Get actor's name
	var actor *tags.Actor
	actorCol := rottensoup.NextSiblingByTag(firstCol, atom.Td)
	if actorCol == nil {
		return nil, errors.New("No actor column found")
	}
	name, err := trimmedCellText(actorCol)
	if err != nil {
		return nil, fmt.Errorf("Could not extract actor's name: %s", err)
	}
	actor = &tags.Actor{Name: name}

	//Get actor's character
	var characterCol *html.Node
	for n := actorCol.NextSibling; n != nil; n = n.NextSibling {
		if n.Data == "td" {
			for _, attr := range n.Attr {
				if attr.Namespace == "" && attr.Key == "class" && attr.Val == "character" {
					characterCol = n
					break
				}
			}
		}
	}
	if characterCol != nil {
		character, err := trimmedCellText(characterCol)
		if err == nil {
			actor.Character = character
		}
	}
	return actor, nil
}

//The table that contains the names for a specific class of cast (directors, writers) except actors
//is preceded by a heading that has an id. This function evaluates the html table that is
//the heading's sibling and extracts all table cells that contain names. These are then
//returned as a slice. Nil will be returned if no names were found.
func (r *Credits) namesByID(id string) []*html.Node {
	heading := rottensoup.ElementByID(r.root, id)
	if heading == nil {
		return nil
	}
	table := rottensoup.NextElementSibling(heading)
	if table.DataAtom != atom.Table {
		return nil
	}
	return rottensoup.ElementsByTagAndAttr(table, atom.Td, html.Attribute{Key: "class", Val: "name"})
}

func (r *Credits) textFromNameCell(n *html.Node) (string, error) {
	link := rottensoup.FirstElementByTag(n, atom.A)
	if link == nil {
		return "", errors.New("No hyperlink found inside table cell")
	}
	text := rottensoup.FirstNodeByType(link, html.TextNode)
	if text == nil {
		return "", errors.New("Hyperlink has no child that is a text node")
	}
	return strings.TrimSpace(text.Data), nil
}
