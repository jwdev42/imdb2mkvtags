//This file is part of imdb2mkvtags ©2022 Jörg Walter

package imdb

import (
	"bytes"
	"fmt"
	"github.com/jwdev42/imdb2mkvtags/internal/global"
	ihttp "github.com/jwdev42/imdb2mkvtags/internal/http"
	"github.com/jwdev42/imdb2mkvtags/internal/lcconv"
	"github.com/jwdev42/rottensoup"
	"golang.org/x/net/html"
	"strconv"
)

type Keyword struct {
	Name  string
	Votes int
}

func ParseKeywordPage(url string, lang *lcconv.LngCntry) ([]Keyword, error) {
	//Fetch document
	body := new(bytes.Buffer)
	if err := ihttp.GetBody(nil, url, body, lang); err != nil {
		return nil, fmt.Errorf("Could not fetch keyword page: %s", err)
	}
	//Parse document
	root, err := html.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("Could not parse keyword page: %s", err)
	}
	table := rottensoup.FirstElementByClassName(root, "dataTable", "evenWidthTable2Col")
	if table == nil {
		return nil, fmt.Errorf("Keyword table not found on keyword page")
	}
	keywordNodes := rottensoup.ElementsByClassName(table, "soda", "sodavote")
	if len(keywordNodes) < 1 {
		return nil, fmt.Errorf("No keywords found on keyword page")
	}
	//Load keyword data from the html node into the keyword struct
	keywords := make([]Keyword, 0, len(keywordNodes))
	for i, node := range keywordNodes {
		kw := Keyword{Votes: -1}
		for _, attr := range node.Attr {
			if attr.Key == "data-item-keyword" && len(attr.Val) > 0 {
				kw.Name = attr.Val
			} else if attr.Key == "data-item-votes" && len(attr.Val) > 0 {
				num, err := strconv.Atoi(attr.Val)
				if err != nil {
					global.Log.Error(fmt.Errorf("Failed to convert attribute data-item-votes to int: %s", err))
				}
				kw.Votes = num
			}
		}
		if len(kw.Name) > 0 {
			keywords = append(keywords, kw)
		} else {
			global.Log.Debug(fmt.Sprintf("ParseKeywordPage: Skipped empty keyword at pos %d", i))
		}
	}
	global.Log.Debug(fmt.Sprintf("ParseKeywordPage: Scraped %d keywords", len(keywords)))
	return keywords, nil
}
