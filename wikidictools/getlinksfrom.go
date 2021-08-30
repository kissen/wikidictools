package wikidictools

import "regexp"

var _DEFINITION_LINK_PATTERN = regexp.MustCompile(`(?m)\[\[(.*?)\]\]`)

// Given defintion, return all [[links]] contained inside it.
func GetLinksFrom(definition string) []string {
	return _DEFINITION_LINK_PATTERN.FindAllString(definition, -1)
}
