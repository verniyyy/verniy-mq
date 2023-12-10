package util

import "strings"

func TrimNullChar(s string) string {
	return strings.Trim(s, "\x00")
}
