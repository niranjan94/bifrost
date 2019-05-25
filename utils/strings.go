package utils

import (
	"strings"
	"unicode/utf8"
)

func ToValidUTF8(value string) string {
	return strings.Map(func(r rune) rune {
		if r == utf8.RuneError {return -1}
		return r
	}, value)
}