package format

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type Table struct {
	title []string
	rows  [][]string
}

func (t *Table) Entitle(headings []string) {
	t.title = headings
}

func (t *Table) AddRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) String() string {
	wds := t.widths()
	bold := color.New(color.Bold).SprintfFunc()
	cyan := color.New(color.FgHiCyan).SprintfFunc()
	result := ""
	for col, tw := range t.title {
		padding := strings.Repeat(" ", wds[col]-len(tw))
		result = result + fmt.Sprintf("%s%s ", bold(tw), padding)
	}

	result = result + "\n"

	for _, r := range t.rows {
		for col, c := range r {
			padding := strings.Repeat(" ", wds[col]-len(c))
			if col == 0 {
				result = result + fmt.Sprintf("%s%s ", cyan(c), padding)
			} else {
				result = result + fmt.Sprintf("%s%s ", c, padding)
			}
		}
		result = result + "\n"
	}

	return result
}

func (t *Table) widths() []int {
	w := []int{}
	for c, _ := range t.title {
		w = append(w, t.width(c))
	}
	return w
}

func (t *Table) width(col int) int {
	width := len(t.title[col])
	for _, r := range t.rows {
		width = Max(width, len(r[col]))
	}
	return width
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
