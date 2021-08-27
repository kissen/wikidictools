package wikidictools

// Return whether this entry is empty in that it contains no defintions.
func (e *DictionaryEntry) IsEmpty() bool {
	return isEmpty(e.Noun) && isEmpty(e.Verb) && isEmpty(e.Adjective) && isEmpty(e.Adverb) && isEmpty(e.Phrase)
}

func isEmpty(slice []string) bool {
	return len(slice) == 0
}
