package wikidictools

import (
	"bufio"
	"github.com/dustin/go-wikiparse"
	"github.com/pkg/errors"
	"io"
	"regexp"
	"strings"
)

// Regex pattern to match {{foo|bar|baz}} inside defintions.
var _META_CURLY_PATTERN = regexp.MustCompile(`(?m){{.*?([^\|]*)}}`)

// Regex pattern to match [[foo|bar|baz]] inside defintions.
var _META_BRACKET_PATTERN = regexp.MustCompile(`(?m)\[\[.*?([^\|]*?)\]\]`)

type xmlParser struct {
	wikiParser wikiparse.Parser
}

// Create new XmlParser for the given stream.
func NewXmlParser(rx io.Reader) (XmlParser, error) {
	wikiParser, err := wikiparse.NewParser(rx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create underlying xml parser")
	}

	return &xmlParser{wikiParser: wikiParser}, nil
}

func (xp *xmlParser) Next() (*DictionaryEntry, error) {
	page, err := nextDictionaryPage(xp.wikiParser)

	if err == io.EOF {
		return nil, err
	}

	if err != nil {
		return nil, errors.Wrap(err, "could not read from underlying parser")
	}

	return pageToDictEntry(page), nil
}

func nextDictionaryPage(parser wikiparse.Parser) (*wikiparse.Page, error) {
	for {
		page, err := parser.Next()

		if err != nil {
			return nil, err
		}

		if !isDictionaryEntry(page) {
			continue
		}

		return page, nil
	}
}

// Return whether the given page appears to be a page related to some specific
// word. Filters out meta pages part of the Wiktionary wiki.
func isDictionaryEntry(page *wikiparse.Page) bool {
	return len(page.Title) > 0 && !strings.ContainsRune(page.Title, ':')
}

type section int

func pageToDictEntry(page *wikiparse.Page) *DictionaryEntry {
	// We are going to fill up this entry line by line.

	entry := DictionaryEntry{
		Word: page.Title,
	}

	scanner := bufio.NewScanner(contentFrom(page))

	// To parse each line, we build up a small DFA with states defined
	// as below.

	const (
		noun section = iota
		verb
		adjective
		adverb
		phrase
		unknown
	)

	inEnglishSection := false
	currentSubSection := unknown

	for scanner.Scan() {
		line := scanner.Text()

		// Check whether this line introduces a change in language/section.

		if isHeading(line) && !inEnglishSection {
			if getLowerHeadingFrom(line) == "english" {
				inEnglishSection = true
				currentSubSection = unknown
			}
		}

		// If we are not changing section but are currently not in the English
		// section, just keep looping. We currently only support the English
		// language.

		if !inEnglishSection {
			continue
		}

		// We are inside the English section. Check whether we found a section
		// that is supported by the DictionaryEntry type.

		if isHeading(line) {
			switch getLowerHeadingFrom(line) {
			case "noun":
				currentSubSection = noun
			case "proper noun":
				currentSubSection = noun
			case "numeral":
				currentSubSection = noun
			case "verb":
				currentSubSection = verb
			case "adjective":
				currentSubSection = adjective
			case "adverb":
				currentSubSection = adverb
			case "phrase":
				currentSubSection = phrase
			default:
				currentSubSection = unknown
			}

			continue
		}

		// If we are in a currently not supported subsection, just keep looping.

		if currentSubSection == unknown {
			continue
		}

		// Now we just add elements for each supported section.

		if isTopLevelListEntry(line) {
			listEntry := getTopLevelListEntryFrom(line)

			switch currentSubSection {
			case noun:
				entry.Noun = append(entry.Noun, listEntry)
			case verb:
				entry.Verb = append(entry.Verb, listEntry)
			case adjective:
				entry.Adjective = append(entry.Adjective, listEntry)
			case adverb:
				entry.Adverb = append(entry.Adverb, listEntry)
			case phrase:
				entry.Phrase = append(entry.Phrase, listEntry)
			}
		}
	}

	return &entry
}

func contentFrom(page *wikiparse.Page) io.Reader {
	latestRevision := &page.Revisions[0]
	return strings.NewReader(latestRevision.Text)
}

func isHeading(line string) bool {
	return strings.HasPrefix(line, "==") && strings.HasSuffix(line, "==")
}

func getLowerHeadingFrom(line string) string {
	return strings.ToLower(strings.Trim(line, "="))
}

func isTopLevelListEntry(line string) bool {
	return listIndentLevel(line) == 1
}

func getTopLevelListEntryFrom(line string) string {
	withoutPrefix := line[1:]
	trimmed := strings.TrimSpace(withoutPrefix)
	withoutCurlies := cleanCurlyBracesFrom(trimmed)
	withoutBrackets := cleanBracketsFrom(withoutCurlies)

	return withoutBrackets
}

func isMediaWikiListChar(r rune) bool {
	const listPrefixChars = "*#;:"
	return strings.ContainsRune(listPrefixChars, r)
}

func listIndentLevel(line string) int {
	for i, r := range line {
		if !isMediaWikiListChar(r) {
			return i
		}
	}

	return len(line)
}

func cleanCurlyBracesFrom(line string) string {
	return _META_CURLY_PATTERN.ReplaceAllString(line, "($1)")
}

func cleanBracketsFrom(line string) string {
	return _META_BRACKET_PATTERN.ReplaceAllString(line, "[[$1]]")
}