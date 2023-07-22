package util

import "strings"

const TagSep = ","

func DecodeTags(tagStr string) []string {
	if tagStr == "" {
		return make([]string, 0)
	}
	return strings.Split(tagStr, TagSep)
}

func EncodeTag(tags []string) string {
	return strings.Join(tags, TagSep)
}
