package htmldriver

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// List represents an HTML list element (ul or ol)
type List struct {
	dom       *DOM
	selection *goquery.Selection
}

// List finds and returns a list element matching the selector
func (d *DOM) List(selector string) (*List, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually a list element
	if !sel.Is("ul") && !sel.Is("ol") {
		return nil, fmt.Errorf("selected element is not a list: %s", selector)
	}

	return &List{
		dom:       d,
		selection: sel,
	}, nil
}

// GetItems returns all list items as a slice of strings
func (l *List) GetItems() ([]string, error) {
	var items []string

	l.selection.Find("li").Each(func(i int, item *goquery.Selection) {
		items = append(items, strings.TrimSpace(item.Text()))
	})

	return items, nil
}

// GetItemCount returns the number of items in the list
func (l *List) GetItemCount() int {
	return l.selection.Find("li").Length()
}
