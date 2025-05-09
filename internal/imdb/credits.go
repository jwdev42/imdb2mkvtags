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
	"regexp"
	"slices"
	"strings"
)

var matchNameLink = regexp.MustCompile("name\\/nm")
var matchCharacterLink = regexp.MustCompile("characters\\/nm")

// represents "fullcredits" pages https://www.imdb.com/title/$titleID/fullcredits
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

	//Get the cast subsection
	table := r.elementByTestID("sub-section-cast")
	if table == nil {
		return nil, errors.New("No cast subsection found")
	}

	actors := make([]tags.Actor, 0, 100)
	rows := rottensoup.ElementsByTag(table, atom.Li)
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

func (r *Credits) Directors() ([]tags.UniLingual, error) {
	return r.scrapeUnilingualSubSection("sub-section-director")
}

func (r *Credits) Producers() ([]tags.UniLingual, error) {
	return r.scrapeUnilingualSubSection("sub-section-producer")
}

func (r *Credits) Writers() ([]tags.UniLingual, error) {
	return r.scrapeUnilingualSubSection("sub-section-writer")
}

func (r *Credits) actor(entry *html.Node) (*tags.Actor, error) {
	getLinkText := func(parent *html.Node) (string, error) {
		text := parent.FirstChild
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
	actorMatches := rottensoup.ElementsByAttrMatch(entry, "", "href", matchNameLink)
	if actorMatches == nil || len(actorMatches) < 2 {
		return nil, errors.New("No actor hyperlink found inside list entry")
	}
	var actor *tags.Actor
	actorName, err := getLinkText(actorMatches[1])
	if err != nil {
		return nil, fmt.Errorf("Failed to extract actor string: %s", err)
	}
	actor = &tags.Actor{Name: actorName}

	//Get actor's character
	characterMatches := rottensoup.ElementsByAttrMatch(entry, "", "href", matchCharacterLink)
	if characterMatches == nil {
		return actor, fmt.Errorf("No character found for actor %q", actorName)
	}
	characterName, err := getLinkText(characterMatches[0])
	if err != nil {
		return actor, fmt.Errorf("Failed to extract character string for actor %q: %s", actorName, err)
	}
	actor.Character = characterName
	return actor, nil
}

func (r *Credits) scrapeUnilingualSubSection(testID string) ([]tags.UniLingual, error) {
	texts, err := r.scrapeTextsFromSubSection(testID)
	if err != nil {
		return nil, err
	}
	uniLingualTexts := make([]tags.UniLingual, len(texts))
	for i, text := range texts {
		uniLingualTexts[i] = tags.UniLingual(text)
	}
	return uniLingualTexts, nil
}

func (r *Credits) scrapeTextsFromSubSection(testID string) ([]string, error) {
	links, err := r.nameLinksFromSubSection(testID)
	if err != nil {
		return nil, err
	}
	texts := r.extractTextChildren(links)
	if texts == nil {
		return nil, fmt.Errorf("Could not extract any text nodes from subsection %q", testID)
	}
	return slices.Compact(texts), nil
}

func (r *Credits) nameLinksFromSubSection(testID string) ([]*html.Node, error) {
	subSection := r.elementByTestID(testID)
	if subSection == nil {
		return nil, fmt.Errorf("No subsection %q found", testID)
	}
	return rottensoup.ElementsByAttrMatch(subSection, "", "href", matchNameLink), nil
}

// Reads the first child of each element in element. If it is a text node,
// its text will be appended to the output slice.
// Returns nil if elements is nil or empty.
func (r *Credits) extractTextChildren(elements []*html.Node) []string {
	if elements == nil || len(elements) == 0 {
		return nil
	}
	texts := make([]string, 0, len(elements))
	for _, element := range elements {
		child := element.FirstChild
		if child == nil {
			continue
		}
		if child.Type == html.TextNode && child.Data != "" {
			texts = append(texts, child.Data)
		}
	}
	return texts
}

func (r *Credits) elementByTestID(testID string) *html.Node {
	return rottensoup.FirstElementByAttr(r.root, html.Attribute{Key: "data-testid", Val: testID})
}
