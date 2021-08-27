package main

import (
	"fmt"
	"github.com/dustin/go-wikiparse"
	"io"
	"os"
	"strings"
)

func isDictionaryEntry(page *wikiparse.Page) bool {
	return !strings.HasPrefix(page.Title, "Wiktionary:")
}

func textFrom(page *wikiparse.Page) string {
	latestRevision := &page.Revisions[0]
	return latestRevision.Text
}

func main() {
	xmlFile := os.Args[1]

	xmlStream, err := os.Open(xmlFile)
	if err != nil {
		panic(err)
	}

	parser, err := wikiparse.NewParser(xmlStream)
	if err != nil {
		panic(err)
	}

	for {
		page, err := parser.Next()

		if err != nil {
			break
		}

		if !isDictionaryEntry(page) {
			continue
		}

		fmt.Println(page.Title, textFrom(page)[:20])
	}

	if err != io.EOF {
			panic(err)
	}
}
