//This file is part of imdb2mkvtags ©2021 Jörg Walter

package lcconv

import (
	"errors"
	"github.com/biter777/countries"
	"github.com/emvi/iso-639-1"
	"regexp"
	"strings"
)

var regexpCountryCode = regexp.MustCompile("^[A-Z]{2}$")

// Provides information about language & country
type LngCntry struct {
	iso6391 string //ISO-639-1 language code
	cc      countries.CountryCode
}

func NewLngCntry(input string) (*LngCntry, error) {
	splitted := strings.Split(input, "-")
	if len(splitted) != 2 {
		return nil, errors.New("Expected language and country separated by \"-\"")
	}
	lng := splitted[0]
	cntry := splitted[1]

	if !iso6391.ValidCode(lng) {
		return nil, errors.New("Invalid language code")
	}

	if !regexpCountryCode.MatchString(cntry) {
		return nil, errors.New("Invalid country code")
	}

	cc := countries.ByName(cntry)
	if cc == countries.Unknown {
		return nil, errors.New("Invalid country code")
	}

	return &LngCntry{
		iso6391: lng,
		cc:      cc,
	}, nil
}

// Returns the 2-character language code.
func (r *LngCntry) ISO6391() string {
	return r.iso6391
}

// Returns the Alpha-2 country code.
func (r *LngCntry) Alpha2() string {
	return r.cc.Alpha2()
}

// Returns the Alpha-3 country code.
func (r *LngCntry) Alpha3() string {
	return r.cc.Alpha3()
}

func (r *LngCntry) HttpHeader() string {
	return r.ISO6391() + "-" + r.Alpha2()
}
