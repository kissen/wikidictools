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

	fmt.Fprintf(os.Stderr, "%v: opened %v for reading\n", os.Args[0], xmlFile)

	// Read entry by entry. Write to db. We create only one big transaction
	// for all entries rather than fine-grained steps. While this might
	// result in problems if the program gets interrupted, doing small transactions
	// drastically reduces performance. So we opt for one big transaction instead.

	nadded := 0

	tx, err := db.Begin()
	if err != nil {
		exitBecauseOf(err)
	}

	for {
		// Get the next dictionary entry.

		entry, err := parser.Next()

		if err != nil {
			break
		}

		if entry.IsEmpty() {
			continue
		}

		if err := InsertDictionaryEntry(tx, entry); err != nil {
			exitBecauseOf(err)
		}

		nadded += 1

		if nadded%1000 == 0 {
			fmt.Fprintf(os.Stderr, "\r%v: processed %v words so far", os.Args[0], nadded)
		}
	}

	if err != nil && err != io.EOF {
		tx.Rollback()
		exitBecauseOf(err)
	}

	if err := tx.Commit(); err != nil {
		exitBecauseOf(err)
	}
}
