package main

import (
	"database/sql"
	"os"

	"github.com/kissen/wikidictools/wikidictools"
	"github.com/pkg/errors"
)

type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

func CreateEmptyFileAt(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		errors.Wrap(err, "could not touch file")
	}

	if err := file.Close(); err != nil {
		errors.Wrap(err, "could not close created file")
	}

	return nil
}

func CreateTablesWith(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.Wrap(err, "could not start transaction")
	}

	if err := createWordTable(tx); err != nil {
		return rollbackBecauseOf(err, tx)
	}

	if err := createDefintionTable(tx); err != nil {
		return rollbackBecauseOf(err, tx)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit failed")
	}

	return nil
}

func InsertDictionaryEntry(tx *sql.Tx, entry *wikidictools.DictionaryEntry) error {
	// First we add the word itself.

	wordId, err := insertWord(tx, entry.Word)
	if err != nil {
		return rollbackBecauseOf(err, tx)
	}

	// Now we add the individual defintions.

	var insertError error

	entry.ForEachDefintion(func(definition string) bool {
		insertError = insertDefintion(tx, wordId, definition)
		return insertError == nil   // keep iterating if no error occured
	})

	if insertError != nil {
		return rollbackBecauseOf(insertError, tx)
	}

	// We are done here! Success!

	return nil
}

// Insert word into the database. Returns the assigned id.
func insertWord(db Preparer, word string) (int64, error) {
	sql := `INSERT INTO words(word) VALUES($1);`
	return insert(db, sql, word)
}

// Insert defintion in the database.
func insertDefintion(db Preparer, wordId int64, defintion string) error {
	sql := `INSERT INTO definitions(word_id, definition) VALUES($1, $2);`
	return execute(db, sql, wordId, defintion)
}

func createWordTable(db Preparer) error {
	sql := `
		CREATE TABLE words (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			word TEXT NO NULL
		);`

	return execute(db, sql)
}

func createDefintionTable(db Preparer) error {
	sql := `
		CREATE TABLE definitions (
			word_id INTEGER,
			definition TEXT NOT NULL,
			FOREIGN KEY(word_id) REFERENCES words(id)
		);`

	return execute(db, sql)
}

func execute(db Preparer, sql string, args... interface{}) error {
	statement, err := db.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "could not prepare statement")
	}

	if _, err := statement.Exec(args...); err != nil {
		return errors.Wrap(err, "could not execute statement")
	}

	return nil
}

func insert(db Preparer, sql string, args... interface{}) (int64, error) {
	statement, err := db.Prepare(sql)
	if err != nil {
		return 0, errors.Wrap(err, "could not prepare statement")
	}

	result, err := statement.Exec(args...)
	if err != nil {
		return 0, errors.Wrap(err, "could not execute statement")
	}

	return result.LastInsertId()
}

// Roll back transaction tx because dbError occured. Returns
// the error that may be passed to higher layers.
func rollbackBecauseOf(dbError error, tx *sql.Tx) error {
	if rollbackError := tx.Rollback(); rollbackError != nil {
		return errors.Wrapf(dbError, "bad rollback: %v: rolled back because:", rollbackError)
	}

	return dbError
}
