//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"github.com/jwdev42/imdb2mkvtags/internal/util"
	"net/url"
	"strings"
)

//Returns true if title is a formally valid imdb title id
func IsTitleID(title string) bool {
	return isValidID("tt", title)
}

func TitleUrl2CreditsUrl(title string) (string, error) {
	u, err := url.Parse(title)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/fullcredits"
	return u.String(), nil
}

func isValidID(prefix, id string) bool {
	if len(id) < 9 {
		return false
	}
	pre := id[:2]
	num := id[2:]
	if pre == prefix && util.IsNumericAscii(num) {
		return true
	}
	return false
}
