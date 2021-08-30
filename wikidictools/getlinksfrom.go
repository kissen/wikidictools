package wikidictools

import (
	"regexp"
	"strings"
)

var _DEFINITION_LINK_PATTERN = regexp.MustCompile(`(?m)\[\[(.*?)\]\]`)

// Given defintion, return all [[links]] contained inside it.
func GetLinksFrom(definition string) (links []string) {
	for _, link := range _DEFINITION_LINK_PATTERN.FindAllString(definition, -1) {
		link = strings.TrimSpace(link)
		link = strings.TrimPrefix(link, "[[")
		link = strings.TrimSuffix(link, "]]")

		links = append(links, link)
	}

	return links
}
