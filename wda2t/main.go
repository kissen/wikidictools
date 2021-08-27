package main

import (
	"fmt"
	"io"
	"os"
	"github.com/kissen/wikidictools/wikidictools"
)

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
		page, err := parser.Next()

		if err != nil {
			break
		}

		fmt.Printf("%v %v %v %v\n", page.Word, len(page.Noun), len(page.Verb), len(page.Adjective))
	}

	if err != io.EOF {
			panic(err)
	}
}
