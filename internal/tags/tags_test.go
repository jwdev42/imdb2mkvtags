//This file is part of imdb2mkvtags ©2021 Jörg Walter

package tags

import (
	"testing"
)

func TestSetField(t *testing.T) {
	const testText = "test"
	m := &Movie{}
	m.SetField("Titles", []MultiLingual{MultiLingual{Text: testText, Lang: "en-US"}})
	if m.Titles[0].Text != testText {
		t.Errorf(`Expected string "%s", got "%s" instead`, testText, m.Titles[0].Text)
	}
}

func TestSetFieldNonexistentField(t *testing.T) {
	defer func() { recover() }()
	m := &Movie{}
	m.SetField("Fnord", []MultiLingual{MultiLingual{Text: "test", Lang: "en-US"}})
	t.Error("Expected panic")
}

func TestSetFieldWrongType(t *testing.T) {
	defer func() { recover() }()
	m := &Movie{}
	m.SetField("Imdb", "blah")
	t.Error("Expected panic")
}
