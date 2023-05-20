//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"errors"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Represents the list containing directors, writers and stars on an imdb title page.
type creditsList map[string][]string

func parseCreditsList(root *html.Node) (creditsList, error) {
	list := make(creditsList)
	section := rottensoup.FirstElementByTagAndAttr(root, atom.Section, html.Attribute{Key: "data-testid", Val: "title-cast"})
	if section == nil {
		return nil, errors.New("CreditsList: No section found with testid 'title-cast'")
	}
	ul := section.FirstChild
	for ul != nil {
		if ul.DataAtom == atom.Ul {
			break
		}
		ul = ul.NextSibling
	}
	if ul == nil {
		return nil, errors.New("CreditsList: No credits list found")
	}
	index := 0
	for node := ul.FirstChild; node != nil; node = node.NextSibling {
		if node.DataAtom != atom.Li {
			continue
		}
		label, err := list.label(index)
		index++
		if err != nil {
			break
		}
		innerList := rottensoup.FirstElementByTag(node, atom.Ul)
		if innerList == nil {
			return nil, fmt.Errorf("CreditsList: No inner list found for label '%s'", label)
		}
		entries := rottensoup.ElementsByTag(innerList, atom.Li)
		if entries == nil {
			return nil, fmt.Errorf("CreditsList: No entries found for label '%s'", label)
		}
		data := make([]string, 0)
		for i, entry := range entries {
			text := rottensoup.FirstNodeByType(entry, html.TextNode)
			if text == nil {
				global.Log.Info(fmt.Errorf("CreditsList: No text at position %d for label '%s'", i, label))
				continue
			}
			data = append(data, text.Data)
		}
		if len(data) < 1 {
			global.Log.Error(fmt.Errorf("CreditsList: No entry in credits list was applicable for label '%s'", label))
		} else {
			list[label] = data
		}
	}
	return list, nil
}

func (r creditsList) label(pos int) (string, error) {
	switch pos {
	case 0:
		return "directors", nil
	case 1:
		return "writers", nil
	}
	return "", fmt.Errorf("No known label for list index %d", pos)
}
