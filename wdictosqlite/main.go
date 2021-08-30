package main

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	"github.com/kissen/wikidictools/wikidictools"

	_ "github.com/mattn/go-sqlite3"
)

func exitBecauseOf(err error) {
	fmt.Fprintf(os.Stderr, "%v: error: %v", os.Args[0], err)
	os.Exit(1)
}

func main() {
	// Parse arguments.

	if len(os.Args) != 1+2 {
		println("usage: wdictosqlite XML_FILE SQLITE_FILE")
		os.Exit(1)
	}

	xmlFile := os.Args[1]
	sqlFile := os.Args[2]

	// Create the SQL file ready for writing.

	if err := CreateEmptyFileAt(sqlFile); err != nil {
		exitBecauseOf(err)
	}

	db, err := sql.Open("sqlite3", sqlFile)
	if err != nil {
		exitBecauseOf(err)
	}

	defer db.Close()

	if err := CreateTablesWith(db); err != nil {
		exitBecauseOf(err)
	}

	fmt.Fprintf(os.Stderr, "%v: created database file %v\n", os.Args[0], sqlFile)

	// Database is now ready. Open the XML file for parsing.

	xmlStream, err := os.Open(xmlFile)
	if err != nil {
		exitBecauseOf(err)
	}

	defer xmlStream.Close()

	parser, err := wikidictools.NewXmlParser(xmlStream)
	if err != nil {
		exitBecauseOf(err)
	}

	fmt.Fprintf(os.Stderr, "%v: opened %v for writing\n", os.Args[0], xmlFile)

	// Read entry by entry. Write to db. We create only one big transaction
	// for all entries rather than fine-grained steps. While this might
	// result in problems if the program gets interrupted, doing small transactions
	// drastically reduces performance. So we opt for one big transaction instead.

	// While we are adding each item to the database, we are also keeping track
	// of the number of references for each individual word. We use that afterwards
	// to compute the number of references.

	nadded := 0
	nreferences := make(map[string]int64)

	tx, err := db.Begin()
	if err != nil {
		exitBecauseOf(err)
	}

	for {
		// Get the next dictionary entry from the parser.

		entry, err := parser.Next()

		if err != nil {
			break
		}

		// Skip words without at least one definition associated with it.

		if entry.IsEmpty() {
			continue
		}

		// Add entry to the database, that is add the (1) word itself and (2) each
		// individual definition.

		if err := InsertDictionaryEntry(tx, entry); err != nil {
			exitBecauseOf(err)
		}

		// Ensure that we are tracking the word in memory.

		addOrIncrement(nreferences, entry.Word, 0)

		// Keep the number of references updated.

		entry.ForEachDefintion(func(definition string) bool {
			for _, link := range wikidictools.GetLinksFrom(definition) {
				addOrIncrement(nreferences, link, 1)
			}

			return true
		})

		// Report on progress.

		nadded += 1

		if nadded%1000 == 0 {
			fmt.Fprintf(os.Stderr, "\r%v: processed %v words for insertion", os.Args[0], nadded)
		}
	}

	if err != nil && err != io.EOF {
		tx.Rollback()
		exitBecauseOf(err)
	}

	if err := tx.Commit(); err != nil {
		exitBecauseOf(err)
	}

	fmt.Fprintf(os.Stderr, "\n%v: done processing %v words for insertion\n", os.Args[0], nadded)

	// Now that we have written all individual words and definitions, we can
	// fill in the nreferences field we kept around. Again we do this in one
	// big transaction.

	ncounted := 0

	tx, err = db.Begin()
	if err != nil {
		exitBecauseOf(err)
	}

	for word, wordReferences := range nreferences {
		// Skip words w/o references. They were already initalized to 0 by
		// the database.

		if wordReferences == 0 {
			continue
		}

		// Set the individual word.

		if err := SetNumberOfReferencesOn(tx, word, wordReferences); err != nil {
			tx.Rollback()
			exitBecauseOf(err)
		}

		// Report on progress.

		ncounted += 1

		if ncounted%1000 == 0 {
			totalToCount := len(nreferences)
			percentageCompleted := float64(ncounted) / float64(totalToCount) * 100.0
			fmt.Fprintf(os.Stderr, "\r%v: set counts for %v (%.1f%%) words", os.Args[0], ncounted, percentageCompleted)
		}
	}

	if err != nil && err != io.EOF {
		tx.Rollback()
		exitBecauseOf(err)
	}

	if err := tx.Commit(); err != nil {
		exitBecauseOf(err)
	}

	fmt.Fprintf(os.Stderr, "\n%v done setting counts for %v words\n", os.Args[0], ncounted)
}

// Increment m[key] by value. If m has no matching key, m[key] is set to value.
func addOrIncrement(m map[string]int64, key string, value int64) {
	if oldValue, ok := m[key]; ok {
		m[key] = oldValue + value
	} else {
		m[key] = value
	}
}
