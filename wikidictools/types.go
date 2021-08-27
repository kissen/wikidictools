package wikidictools

// Represents a parser to MediaWiki XML exports.
type XmlParser interface {
	// Return the next entry from the given dictionary. On success, retrns a
	// non-nil dictionary entry and a nil error. If the reading has caused an
	// error, that error is returned as-is.  In particular, if end of file was
	// reached, this method returns (nil, io.EOF) which in most cases is not a
	// failure case.
	Next() (*DictionaryEntry, error)
}

// A single entry of the dictionary.
type DictionaryEntry struct {
	// Noun defintions. Each entry in the slice contains one possible defintion.
	// May be nil.
	Noun []string

	// Noun defintions. Each entry in the slice contains one possible defintion.
	// May be nil.
	Verb []string

	// Adjective defintions. Each entry in the slice contains one possible
	// defintion. May be nil.
	Adjective []string
}
