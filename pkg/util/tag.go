package util

import (
	"strings"
)

// k8s不让用逗号，麻了...
const TagSep = "."

func DecodeTags(tagStr string) []string {
	if tagStr == "" {
		//panic("")
		return make([]string, 0)
	}

	//log.Println(strings.Split(tagStr, TagSep))
	return strings.Split(tagStr, TagSep)
}

func EncodeTag(tags []string) string {
	return strings.Join(tags, TagSep)
}
