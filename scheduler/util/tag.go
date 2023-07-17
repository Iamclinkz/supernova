package util

import "strings"

const TagSep = ","

func DecodeTags(tagStr string) []string {
	return strings.Split(tagStr, TagSep)
}

func EncodeTag(tags []string) string {
	return strings.Join(tags, TagSep)
}
