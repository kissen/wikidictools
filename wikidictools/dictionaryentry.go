package wikidictools

// Return whether this entry is empty in that it contains no defintions.
func (e *DictionaryEntry) IsEmpty() bool {
	return isEmpty(e.Noun) && isEmpty(e.Verb) && isEmpty(e.Adjective) && isEmpty(e.Adverb) && isEmpty(e.Phrase)
}

// Run function f on each defintion contained in this dictionary entry.
// Function f returns a bool. If f returns true, ForEachDefintion keeps
// iterating. If f returns false, iteration stops.
func (e *DictionaryEntry) ForEachDefintion(f func(string) bool) {
	choices := [][]string{
		e.Noun, e.Verb, e.Adjective, e.Adverb, e.Phrase,
	}

	for _, choice := range choices {
		for _, word := range choice {
			if !f(word) {
				return
			}
		}
	}
}

func isEmpty(slice []string) bool {
	return len(slice) == 0
}
