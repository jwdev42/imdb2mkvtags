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
	"golang.org/x/net/html/atom"
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
	keywordNodes := rottensoup.ElementsByTagAndAttr(root, atom.Li,
		html.Attribute{Key: "data-testid", Val: "list-summary-item"})
	if len(keywordNodes) < 1 {
		return nil, fmt.Errorf("No keywords found on keyword page")
	}
	//Load keyword data from the html node into the keyword struct
	keywords := make([]Keyword, 0, len(keywordNodes))
	for i, node := range keywordNodes {
		kw := Keyword{Votes: -1}
		//Fill keyword text
		nameNode := rottensoup.FirstElementByTagAndAttr(node, atom.A,
			html.Attribute{Key: "class", Val: "ipc-metadata-list-summary-item__t"})
		if nameNode == nil {
			global.Log.Errorf("No keyword node found for element %d in keyword list", i+1)
			continue
		}
		nameTextNode := rottensoup.FirstNodeByType(nameNode, html.TextNode)
		if nameTextNode == nil {
			global.Log.Errorf("No keyword text node found for element %d in keyword list", i+1)
			continue
		}
		if len(nameTextNode.Data) < 1 {
			global.Log.Errorf("Empty keyword text found for element %d in keyword list", i+1)
			continue
		}
		kw.Name = nameTextNode.Data
		//Votes cannot be evaluated anymore as the data is now generated in the browser via javascript
		//Add keyword to keyword list
		keywords = append(keywords, kw)
	}
	global.Log.Debugf("ParseKeywordPage: Scraped %d keywords", len(keywords))
	return keywords, nil
}
