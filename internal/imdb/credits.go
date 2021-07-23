package imdb

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/tags"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"os"
	"strings"
)

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

func (r *Credits) Cast() ([]tags.Actor, error) {

	//Get the cast table
	var table *html.Node
	{
		tables := rottensoup.ElementsByTagAndAttr(r.root, "table", html.Attribute{Key: "class", Val: "cast_list"})
		if len(tables) < 1 {
			return nil, errors.New("No cast table found")
		}
		table = tables[0]
	}

	actors := make([]tags.Actor, 0, 10)
	rows := rottensoup.ElementsByTagAndAttr(table, "td", html.Attribute{Key: "class", Val: "primary_photo"})
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
