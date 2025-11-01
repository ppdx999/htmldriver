package htmldriver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Link represents an HTML anchor element
type Link struct {
	dom       *DOM
	selection *goquery.Selection
}

// Link finds and returns a link element matching the selector
func (d *DOM) Link(selector string) (*Link, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually a link element
	if !sel.Is("a") {
		return nil, fmt.Errorf("selected element is not a link: %s", selector)
	}

	return &Link{
		dom:       d,
		selection: sel,
	}, nil
}

// GetURL returns the URL of the link
func (l *Link) GetURL() (*url.URL, error) {
	href, exists := l.selection.Attr("href")
	if !exists {
		return nil, fmt.Errorf("link has no href attribute")
	}

	linkURL, err := url.Parse(href)
	if err != nil {
		return nil, fmt.Errorf("invalid link URL: %w", err)
	}

	// Make relative URLs absolute
	if !linkURL.IsAbs() {
		linkURL = l.dom.baseURL.ResolveReference(linkURL)
	}

	return linkURL, nil
}

// GetText returns the text content of the link
func (l *Link) GetText() (string, error) {
	return strings.TrimSpace(l.selection.Text()), nil
}

// Click follows the link and returns the response
func (l *Link) Click() (Response, error) {
	linkURL, err := l.GetURL()
	if err != nil {
		return Response{}, err
	}

	// Build request
	req := l.dom.buildRequest(http.MethodGet, linkURL)

	// Send request
	resp, err := l.dom.transport.Do(req)
	if err != nil {
		return Response{}, err
	}

	// Update cookies from response
	l.dom.updateCookies(resp)

	return resp, nil
}

// Button represents an HTML button element
type Button struct {
	dom       *DOM
	selection *goquery.Selection
}

// Button finds and returns a button element matching the selector
func (d *DOM) Button(selector string) (*Button, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually a button element or input[type=button/submit]
	if !sel.Is("button") && !sel.Is("input[type=\"button\"]") && !sel.Is("input[type=\"submit\"]") {
		return nil, fmt.Errorf("selected element is not a button: %s", selector)
	}

	return &Button{
		dom:       d,
		selection: sel,
	}, nil
}

// GetText returns the text content of the button
func (b *Button) GetText() (string, error) {
	if b.selection.Is("input") {
		// For input elements, use the value attribute
		value, exists := b.selection.Attr("value")
		if !exists {
			return "", nil
		}
		return value, nil
	}
	// For button elements, use text content
	return strings.TrimSpace(b.selection.Text()), nil
}

// Img represents an HTML img element
type Img struct {
	dom       *DOM
	selection *goquery.Selection
}

// Img finds and returns an img element matching the selector
func (d *DOM) Img(selector string) (*Img, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return nil, err
	}

	// Verify it's actually an img element
	if !sel.Is("img") {
		return nil, fmt.Errorf("selected element is not an image: %s", selector)
	}

	return &Img{
		dom:       d,
		selection: sel,
	}, nil
}

// GetSrc returns the src attribute of the image
func (i *Img) GetSrc() (string, error) {
	src, exists := i.selection.Attr("src")
	if !exists {
		return "", fmt.Errorf("image has no src attribute")
	}
	return src, nil
}

// GetAlt returns the alt attribute of the image
func (i *Img) GetAlt() (string, error) {
	alt, exists := i.selection.Attr("alt")
	if !exists {
		return "", fmt.Errorf("image has no alt attribute")
	}
	return alt, nil
}
