package main

import (
	"fmt"
	"github.com/kissen/wikidictools/wikidictools"
	"io"
	"os"
)

func printAllOf(entry *wikidictools.DictionaryEntry) {
	fmt.Println(entry.Word)

	for i, noun := range entry.Noun {
		fmt.Printf("  %v. %v\n", i+1, noun)
	}
}

func main() {
	xmlFile := os.Args[1]

	xmlStream, err := os.Open(xmlFile)
	if err != nil {
		panic(err)
	}

	defer xmlStream.Close()

	parser, err := wikidictools.NewXmlParser(xmlStream)
	if err != nil {
		panic(err)
	}

	for {
		entry, err := parser.Next()

		if err != nil {
			break
		}

		if !entry.IsEmpty() {
			printAllOf(entry)
		}
	}

	if false && err != io.EOF {
		panic(err)
	}
}
