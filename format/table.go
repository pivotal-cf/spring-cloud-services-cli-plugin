/*
 * Copyright (C) 2016-Present Pivotal Software, Inc. All rights reserved.
 *
 * This program and the accompanying materials are made available under
 * the terms of the under the Apache License, Version 2.0 (the "License”);
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
		result += fmt.Sprintf("%s%s ", bold(tw), padding)
	}

	result += "\n"

	for _, r := range t.rows {
		for col, c := range r {
			padding := strings.Repeat(" ", wds[col]-len(c))
			if col == 0 {
				result += fmt.Sprintf("%s%s ", cyan(c), padding)
			} else {
				result += fmt.Sprintf("%s%s ", c, padding)
			}
		}
		result += "\n"
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
