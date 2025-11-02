package htmldriver

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Element represents a generic HTML element
type Element struct {
	selection *goquery.Selection
}

// Text returns the text content of the element
func (e *Element) Text() string {
	return strings.TrimSpace(e.selection.Text())
}

// Attr returns the value of the specified attribute
// Returns empty string if attribute doesn't exist
func (e *Element) Attr(name string) string {
	value, _ := e.selection.Attr(name)
	return value
}

// HasAttr returns true if the element has the specified attribute
func (e *Element) HasAttr(name string) bool {
	_, exists := e.selection.Attr(name)
	return exists
}

// HTML returns the HTML content of the element
func (e *Element) HTML() (string, error) {
	return e.selection.Html()
}

// Is checks if the element matches the given selector
func (e *Element) Is(selector string) bool {
	return e.selection.Is(selector)
}
