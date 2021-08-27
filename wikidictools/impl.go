package wikidictools

import (
	"bufio"
	"github.com/dustin/go-wikiparse"
	"github.com/pkg/errors"
	"io"
	"strings"
)

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

func isDictionaryEntry(page *wikiparse.Page) bool {
	return !strings.HasPrefix(page.Title, "Wiktionary:")
}

type section int

const (
	noun section = iota
	verb
	adjective
	unknown
)

func pageToDictEntry(page *wikiparse.Page) *DictionaryEntry {
	var entry DictionaryEntry
	scanner := bufio.NewScanner(contentFrom(page))

	inEnglishSection := false
	currentSubSection := unknown

	for scanner.Scan() {
		line := scanner.Text()

		// Check whether this line introduces a change in language/section.

		if isH2(line) {
			inEnglishSection = getH2From(line) == "English"
			continue
		}

		// If we are not changing section but are currently not in the English
		// section, just keep looping. We currently only support the English
		// language.

		if !inEnglishSection {
			continue
		}

		// We are inside the English section. Check whether we found a section
		// that is supported by the DictionaryEntry type.

		if isH3(line) {
			switch getH3From(line) {
			case "Noun":
				currentSubSection = noun
			case "Verb":
				currentSubSection = verb
			case "Adjective":
				currentSubSection = adjective
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
			}
		}
	}

	return &entry
}

func contentFrom(page *wikiparse.Page) io.Reader {
	latestRevision := &page.Revisions[0]
	return strings.NewReader(latestRevision.Text)
}

func isH2(line string) bool {
	startOk := strings.HasPrefix(line, "==") && !strings.HasPrefix(line, "===")
	endOk := strings.HasSuffix(line, "==") && !strings.HasSuffix(line, "===")

	return startOk && endOk
}

func getH2From(line string) string {
	return strings.ReplaceAll(line, "==", "")
}

func isH3(line string) bool {
	return strings.HasPrefix(line, "===") && strings.HasSuffix(line, "===")
}

func getH3From(line string) string {
	return strings.ReplaceAll(line, "===", "")
}

func isTopLevelListEntry(line string) bool {
	return listIndentLevel(line) == 1
}

func getTopLevelListEntryFrom(line string) string {
	withoutPrefix := line[1:]
	return strings.TrimSpace(withoutPrefix)
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
