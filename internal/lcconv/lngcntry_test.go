//This file is part of imdb2mkvtags ©2021 Jörg Walter

package lcconv

import (
	"github.com/biter777/countries"
	"testing"
)

func TestNewLngCntry(t *testing.T) {
	de, err := NewLngCntry("de-DE")
	if err != nil {
		t.Error(err)
	} else {
		if de.cc != countries.Germany {
			t.Error("Expected Germany")
		}
	}
	if res := de.Alpha2(); res != "DE" {
		t.Errorf("Expected \"DE\", got \"%s\"", res)
	}
	if res := de.Alpha3(); res != "DEU" {
		t.Errorf("Expected \"DEU\", got \"%s\"", res)
	}

	fails := []string{"DE-de", "en-Us", "En-US", "en-us", "de CH", "en_CA", "a-B", "abc-def", "ab-c", "de-Ö", "ab-CD-ef"}

	for _, fail := range fails {
		_, err := NewLngCntry(fail)
		if err == nil {
			t.Errorf("Expected an error return when calling NewLngCntry with input string \"%s\"", fail)
		}
	}
}
