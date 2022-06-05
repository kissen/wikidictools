package wikidictools

import "io"

// Represents a parser to MediaWiki XML exports.
type XmlParser interface {
	// Return the next entry from the given dictionary. On success, retrns a
	// non-nil dictionary entry and a nil error. If the reading has caused an
	// error, that error is returned as-is.  In particular, if end of file was
	// reached, this method returns (nil, io.EOF) which in most cases is not a
	// failure case.
	Next() (*DictionaryEntry, error)

	io.Closer
}

// A single entry of the dictionary.
type DictionaryEntry struct {
	// Word this entry is about.
	Word string

	// Revision of this particular Wiktionary page.
	Revision uint64

	// Noun defintions. Each entry in the slice contains one possible defintion.
	// May be nil.
	Noun []string

	// Noun defintions. Each entry in the slice contains one possible defintion.
	// May be nil.
	Verb []string

	// Adjective defintions. Each entry in the slice contains one possible
	// defintion. May be nil.
	Adjective []string

	// Adverb defintions. Each entry in the slice contains one possible
	// defintion. May be nil.
	Adverb []string

	// Phrase defintions. Each entry in the slice contains one possible
	// defintion. May be nil.
	Phrase []string
}
