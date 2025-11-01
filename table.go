package htmldriver

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Table represents an HTML table element
type Table struct {
	dom       *DOM
	selection *goquery.Selection
}

// Table finds and returns a table element matching the selector
func (d *DOM) Table(selector string) (*Table, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually a table element
	if !sel.Is("table") {
		return nil, fmt.Errorf("selected element is not a table: %s", selector)
	}

	return &Table{
		dom:       d,
		selection: sel,
	}, nil
}

// GetRows returns all rows in the table as a 2D string slice
// Each row is represented as a slice of cell contents
func (t *Table) GetRows() ([][]string, error) {
	var rows [][]string

	// Find all rows (both in thead and tbody)
	t.selection.Find("tr").Each(func(i int, row *goquery.Selection) {
		var cells []string
		// Get both th and td cells
		row.Find("th, td").Each(func(j int, cell *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(cell.Text()))
		})
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	})

	return rows, nil
}

// GetRowCount returns the number of rows in the table
func (t *Table) GetRowCount() int {
	return t.selection.Find("tr").Length()
}

// GetColCount returns the number of columns in the first row
func (t *Table) GetColCount() int {
	firstRow := t.selection.Find("tr").First()
	return firstRow.Find("th, td").Length()
}
