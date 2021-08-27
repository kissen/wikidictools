package main

import (
	"github.com/kissen/wikidictools/wikidictools"
	"io"
	"os"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Parse arguments.

	if len(os.Args) != 1 + 2 {
		println("usage: wdictosqlite XML_FILE SQLITE_FILE")
		os.Exit(1)
	}

	xmlFile := os.Args[1]
	sqlFile := os.Args[2]

	// Create the SQL file ready for writing.

	if err := CreateEmptyFileAt(sqlFile); err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", sqlFile)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := CreateTablesWith(db); err != nil {
		panic(err)
	}

	// Database is now ready. Open the XML file for parsing.

	xmlStream, err := os.Open(xmlFile)
	if err != nil {
		panic(err)
	}

	defer xmlStream.Close()

	parser, err := wikidictools.NewXmlParser(xmlStream)
	if err != nil {
		panic(err)
	}

	// Read entry by entry. Write to db.

	nadded := 0

	for {
		entry, err := parser.Next()

		if err != nil {
			break
		}

		println("INSERT", entry.Word)

		if err := InsertDictionaryEntry(db, entry); err != nil {
			panic(err)
		}

		nadded += 1
		if nadded >= 100 {
			break
		}
	}

	if false && err != io.EOF {
		panic(err)
	}
}
