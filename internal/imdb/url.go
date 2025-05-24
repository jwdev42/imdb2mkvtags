//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"github.com/jwdev42/imdb2mkvtags/internal/util"
)

// Returns true if title is a formally valid imdb title id
func IsTitleID(title string) bool {
	return isValidID("tt", title)
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
