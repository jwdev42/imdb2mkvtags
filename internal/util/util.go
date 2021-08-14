//This file is part of imdb2mkvtags ©2021 Jörg Walter

package util

func IsNumericAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		b := byte(s[i])
		if b < 0x30 || b > 0x39 {
			return false
		}
	}
	return true
}
