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

//Represents the list containing directors, writers and stars on an imdb title page.
type creditsList map[string][]string

func parseCreditsList(root *html.Node) (creditsList, error) {
	list := make(creditsList)

	div := rottensoup.FirstElementByTagAndAttr(root, atom.Div, html.Attribute{Key: "data-testid", Val: "title-pc-wide-screen"})
	if div == nil {
		return nil, errors.New("CreditsList: No div container found with attribute data-testid=title-pc-wide-screen")
	}
	ul := rottensoup.FirstElementByTag(div, atom.Ul)
	if ul == nil {
		return nil, errors.New("CreditsList: No credits list found")
	}

	entries := rottensoup.ElementsByTagAndAttr(ul, atom.Li, html.Attribute{Key: "data-testid", Val: "title-pc-principal-credit"})
	if entries == nil {
		return nil, errors.New("CreditsList: No entries found in credits list")
	}
	for i, entry := range entries {
		label, err := list.label(i)
		if err != nil {
			global.Log.Error(fmt.Errorf("CreditsList: %s", err))
		}
		global.Log.Debug(fmt.Sprintf("CreditsList: Found label \"%s\"", label))

		//Extract inline data
		inlineNodes := rottensoup.ElementsByClassName(entry, "ipc-inline-list__item")
		if inlineNodes == nil {
			global.Log.Error(fmt.Errorf("CreditsList: No inline list elements found for entry \"%s\" in credits list", label))
			continue
		}
		data := make([]string, 0)
		for j, inlineNode := range inlineNodes {
			a := rottensoup.FirstElementByTag(inlineNode, atom.A)
			if a == nil {
				global.Log.Error(fmt.Errorf("CreditsList: No hyperlink element found in inline list entry %d in credits list entry \"%s\"", j, label))
				continue
			}
			text := rottensoup.FirstNodeByType(a, html.TextNode)
			if text == nil || len(text.Data) < 1 {
				global.Log.Error(fmt.Errorf("CreditsList: No text node found in inline list entry %d in credits list entry \"%s\"", j, label))
				continue
			}
			data = append(data, text.Data)
		}
		if len(data) < 1 {
			global.Log.Error(fmt.Errorf("CreditsList: No data found for entry \"%s\" in credits list", label))
			continue
		}
		list[label] = data
	}
	if len(list) < 1 {
		return nil, fmt.Errorf("CreditsList: No entry in credits list was applicable")
	}
	return list, nil
}

func (r creditsList) label(pos int) (string, error) {
	switch pos {
	case 0:
		return "directors", nil
	case 1:
		return "writers", nil
	case 2:
		return "actors", nil
	}
	return "", fmt.Errorf("No known label for list index %d", pos)
}
