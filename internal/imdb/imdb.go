//This file is part of imdb2mkvtags ©2021 Jörg Walter

package imdb

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

//Convert an IMDB ID string to its corresponding https URL string
func Id2Url(id string) (string, error) {
	const errMsg = "Malformed IMDB ID"
	isNumeric := func(s string) bool {
		for i := 0; i < len(s); i++ {
			b := byte(s[i])
			if b < 0x30 || b > 0x39 {
				return false
			}
		}
		return true
	}
	if len(id) < 9 {
		return "", errors.New(errMsg)
	}
	prefix := id[:2]
	num := id[2:]
	if !isNumeric(num) {
		return "", errors.New(errMsg)
	}

	switch prefix {
	//sorted alphabetically
	case "co":
		return fmt.Sprintf("https://www.imdb.com/company/%s/", id), nil
	case "ls":
		return fmt.Sprintf("https://www.imdb.com/list/%s/", id), nil
	case "nm":
		return fmt.Sprintf("https://www.imdb.com/name/%s/", id), nil
	case "rw":
		return fmt.Sprintf("https://www.imdb.com/review/%s/", id), nil
	case "tt":
		return fmt.Sprintf("https://www.imdb.com/title/%s/", id), nil
	case "ur":
		return fmt.Sprintf("https://www.imdb.com/user/%s/", id), nil
	case "vi":
		return fmt.Sprintf("https://www.imdb.com/video/%s/", id), nil
	default:
		return "", fmt.Errorf("%s or unsupported prefix", errMsg)
	}
}

func TitleUrl2CreditsUrl(title string) (string, error) {
	u, err := url.Parse(title)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/fullcredits"
	return u.String(), nil
}
