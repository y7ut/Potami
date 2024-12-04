package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
	"golang.org/x/text/width"
)

type Table struct {
	Headers []string
	Rows    [][]string
	Width   int
}

func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    make([][]string, 0),
		Width:   len(headers) + 1,
	}
}

func (t *Table) AddRow(row []string) {
	t.Rows = append(t.Rows, row)
}

func (t *Table) Render() {
	currentWidth, err := getTerminalWidth()
	if err != nil {
		log.Fatal(err)
	}

	headersLine := ""
	for i := 0; i < len(t.Headers); i++ {
		headersLine += t.Headers[i]
		if i < len(t.Headers)-1 {
			blockSize := currentWidth/t.Width - getStringWidth(t.Headers[i])
			if blockSize < 0 {
				blockSize = 0
			}
			headersLine += strings.Repeat(" ", blockSize)
		}
	}
	fmt.Println(headersLine)

	for _, row := range t.Rows {
		rowline := ""
		for i := 0; i < len(row); i++ {
			value := row[i]
			if currentWidth/len(row) < getStringWidth(value) {
				value = editStringSlim(value, currentWidth/t.Width)
			}
			rowline += value
			if i < len(row)-1 {
				blockSize := currentWidth/t.Width - getStringWidth(value)
				if blockSize < 0 {
					blockSize = 0
				}
				rowline += strings.Repeat(" ", blockSize)
			}
		}
		fmt.Println(rowline)
	}
}

func getStringWidth(s string) int {
	w := 0
	for _, r := range s {
		switch width.LookupRune(r).Kind() {
		case width.EastAsianFullwidth, width.EastAsianWide:
			w += 2
		case width.EastAsianHalfwidth, width.EastAsianNarrow,
			width.Neutral, width.EastAsianAmbiguous:
			w += 1
		}
	}
	return w
}

func editStringSlim(s string, w int) string {
	w = w - 4
	if getStringWidth(s) < w {
		return s
	}
	for i := len(s) - 1; i >= 0; i-- {
		if getStringWidth(s[:i]) < w {
			if i%3 != 0 {
				i = i - (i % 3)
			}
			return s[:i] + "... "
		}
	}
	return s
}

func getTerminalWidth() (int, error) {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 0, err
	}
	return width, nil
}
