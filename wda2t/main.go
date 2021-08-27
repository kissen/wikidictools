package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

type Page struct {
	Title string `xml:"title"`
	Ns int64 `xml:"ns"`
	Id int64 `xml:"id"`
	Format string `xml:"format"`
	Text string `xml:"revision>text"`
}

func (p *Page) String() string {
	return fmt.Sprintf("{Title=%v Ns=%v Id=%v Format=%v Text=#%v}", p.Title, p.Ns, p.Id, p.Format, len(p.Text))
}

func isDictionaryEntry(p *Page) bool {
	return !strings.HasPrefix(p.Title, "Wiktionary:")
}

func main() {
	xmlFile := os.Args[1]
	xmlStream, err := os.Open(xmlFile)
	if err != nil {
		panic(err)
	}

	decoder := xml.NewDecoder(xmlStream)

	for {
		abstractToken, err := decoder.Token()

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}

		var page Page

		switch token := abstractToken.(type) {
		case xml.StartElement:
			if token.Name.Local == "page" {
				if err := decoder.DecodeElement(&page, &token); err != nil {
					panic(err)
				}

				if !isDictionaryEntry(&page) {
					continue
				}

				fmt.Println(&page)
			}
		}
	}

	fmt.Println("done")
}
