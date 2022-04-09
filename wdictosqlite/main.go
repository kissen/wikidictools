package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kissen/wikidictools/wikidictools"
	"github.com/pkg/errors"

	_ "github.com/mattn/go-sqlite3"
)

type Arguments struct {
	XmlFile   string
	SqlFile   string
	TimeStamp string
}

type ReferencesMap map[string]int64

func ParseArguments() Arguments {
	var (
		args       Arguments
		printUsage bool
	)

	flag.StringVar(&args.XmlFile, "infile", "--", "file from which to read XML or -- for stdin")
	flag.StringVar(&args.SqlFile, "outfile", "", "file to write to, required")

	now := time.Now().UTC().Format("2006-01-02T15:04:05")
	flag.StringVar(&args.TimeStamp, "timestamp", now, "overwrite timestamp embedded in created database")

	flag.BoolVar(&printUsage, "help", false, "print help")

	flag.Parse()

	if args.SqlFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	return args
}

func CreateDatabaseFile(sqlFile string) error {
	if err := CreateEmptyFileAt(sqlFile); err != nil {
		exitBecauseOf(err)
	}

	db, err := sql.Open("sqlite3", sqlFile)
	if err != nil {
		return errors.Wrap(err, "could not create database")
	}

	defer db.Close()

	if err := CreateTablesWith(db); err != nil {
		return errors.Wrap(err, "could not create tables")
	}

	fmt.Fprintf(os.Stderr, "%v: created database file %v\n", os.Args[0], sqlFile)
	return nil
}

func OpenInputFileFrom(fileLocation string) (wikidictools.XmlParser, error) {
	var openedAFile bool
	var rx io.ReadCloser

	switch fileLocation {
	case "--":
		rx = os.Stdin
		openedAFile = false
		fmt.Fprintf(os.Stderr, "%v: using stdin for reading\n", os.Args[0])
	default:
		fd, err := os.Open(fileLocation)
		if err != nil {
			return nil, errors.Wrap(err, "could not open output file")
		}

		rx = fd
		openedAFile = true
		fmt.Fprintf(os.Stderr, "%v: opened %v for reading\n", os.Args[0], fileLocation)
	}

	parser, err := wikidictools.NewXmlParser(rx)
	if err != nil {
		if openedAFile {
			rx.Close()
		}
		return nil, errors.Wrap(err, "could not create wiktionary parser")
	}

	return parser, nil
}

func FillDatabase(dst *sql.DB, src wikidictools.XmlParser) (ReferencesMap, error) {
	nadded := 0
	nreferences := make(ReferencesMap)

	tx, err := dst.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "could not create transaction")
	}

	defer tx.Rollback()

	for {
		// Get the next dictionary entry from the parser.

		entry, err := src.Next()

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
			return nil, errors.Wrapf(err, "could not add entry for word=%v", entry.Word)
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
		return nil, errors.Wrap(err, "error while getting next XML entry")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "could not commit")
	}

	fmt.Fprintf(os.Stderr, "\n%v: done processing %v words for insertion\n", os.Args[0], nadded)
	return nreferences, nil
}

func FillInReferences(dst *sql.DB, nreferences ReferencesMap) error {
	// Now that we have written all individual words and definitions, we can
	// fill in the nreferences field we kept around. Again we do this in one
	// big transaction.

	ncounted := 0

	tx, err := dst.Begin()
	if err != nil {
		return errors.Wrap(err, "could not start transaction")
	}

	defer tx.Rollback()

	for word, wordReferences := range nreferences {
		// Update number of looked at words. Report progress.

		ncounted += 1

		if ncounted%1000 == 0 {
			totalToCount := len(nreferences)
			percentageCompleted := float64(ncounted) / float64(totalToCount) * 100.0
			fmt.Fprintf(os.Stderr, "\r%v: set counts for %v (%.1f%%) words", os.Args[0], ncounted, percentageCompleted)
		}

		// Skip words w/o references. They were already initalized to 0 by
		// the database.

		if wordReferences == 0 {
			continue
		}

		// Set the individual word.

		if err := SetNumberOfReferencesOn(tx, word, wordReferences); err != nil {
			return errors.Wrapf(err, "could not set nreferences on word=%v", word)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "could not commit")
	}

	fmt.Fprintf(os.Stderr, "\n%v done setting counts for %v words\n", os.Args[0], ncounted)
	return nil
}

func WriteMetaData(dst *sql.DB, args *Arguments) error {
	if err := InsertMeta(dst, "TimeStamp", args.TimeStamp); err != nil {
		return err
	}

	return nil
}

func exitBecauseOf(err error) {
	fmt.Fprintf(os.Stderr, "%v: error: %v", os.Args[0], err)
	os.Exit(1)
}

// Increment m[key] by value. If m has no matching key, m[key] is set to value.
func addOrIncrement(m map[string]int64, key string, value int64) {
	if oldValue, ok := m[key]; ok {
		m[key] = oldValue + value
	} else {
		m[key] = value
	}
}

func main() {
	// Parse arguments.

	args := ParseArguments()

	// Truncate and initalize schema in DB file.

	if err := CreateDatabaseFile(args.SqlFile); err != nil {
		exitBecauseOf(err)
	}

	// Start reading XML.

	xmlStream, err := OpenInputFileFrom(args.XmlFile)
	if err != nil {
		exitBecauseOf(err)
	}

	defer xmlStream.Close()

	// Prepare database connection we will use throughout
	// this operation.

	db, err := sql.Open("sqlite3", args.SqlFile)
	if err != nil {
		exitBecauseOf(err)
	}

	defer db.Close()

	// Fill the database. This is where most work gets done.

	nreferences, err := FillDatabase(db, xmlStream)
	if err != nil {
		exitBecauseOf(err)
	}

	if err := FillInReferences(db, nreferences); err != nil {
		exitBecauseOf(err)
	}

	if err := WriteMetaData(db, &args); err != nil {
		exitBecauseOf(err)
	}
}
