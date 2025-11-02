package htmldriver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// CookieJar stores cookies across requests
type CookieJar struct {
	cookies map[string]string
}

// NewCookieJar creates a new cookie jar
func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make(map[string]string),
	}
}

// DOM represents a parsed HTML document
type DOM struct {
	doc       *goquery.Document
	transport Transport
	baseURL   *url.URL
	cookieJar *CookieJar
}

// New creates a new DOM instance with the given transport
// Cookies are NOT shared between DOM instances by default
func New(transport Transport) *DOM {
	// Default base URL
	defaultURL, _ := url.Parse("http://localhost")
	return &DOM{
		transport: transport,
		baseURL:   defaultURL,
		cookieJar: NewCookieJar(),
	}
}

// NewWithCookieJar creates a new DOM instance with a shared cookie jar
// Use this to share cookies across multiple DOM instances
func NewWithCookieJar(transport Transport, jar *CookieJar) *DOM {
	defaultURL, _ := url.Parse("http://localhost")
	return &DOM{
		transport: transport,
		baseURL:   defaultURL,
		cookieJar: jar,
	}
}

// Parse parses the given HTML string and returns the DOM
func (d *DOM) Parse(html string) *DOM {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		// In practice, goquery rarely fails on valid HTML strings
		// but we keep the same instance and set a nil doc if it happens
		d.doc = nil
		return d
	}
	d.doc = doc
	return d
}

// SetCookie sets a cookie that will be sent with subsequent requests
func (d *DOM) SetCookie(name, value string) *DOM {
	d.cookieJar.cookies[name] = value
	return d
}

// GetCookieJar returns the cookie jar for sharing across DOM instances
func (d *DOM) GetCookieJar() *CookieJar {
	return d.cookieJar
}

// SetBaseURL sets the base URL for resolving relative URLs
func (d *DOM) SetBaseURL(baseURL string) (*DOM, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return d, fmt.Errorf("invalid base URL: %w", err)
	}
	d.baseURL = u
	return d, nil
}

// resolveSelector converts custom selector syntax to goquery selector
// @xxxxx -> [test-id="xxxxx"]
// #xxxxx -> #xxxxx (unchanged)
func (d *DOM) resolveSelector(selector string) string {
	if strings.HasPrefix(selector, "@") {
		testID := strings.TrimPrefix(selector, "@")
		return fmt.Sprintf("[test-id=\"%s\"]", testID)
	}
	return selector
}

// findOne finds a single element matching the selector
func (d *DOM) findOne(selector string) (*goquery.Selection, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document not parsed")
	}

	resolved := d.resolveSelector(selector)
	sel := d.doc.Find(resolved)

	if sel.Length() == 0 {
		return nil, fmt.Errorf("element not found: %s", selector)
	}

	if sel.Length() > 1 {
		return nil, fmt.Errorf("multiple elements found for selector: %s", selector)
	}

	return sel, nil
}

// buildRequest creates a Request with common headers and cookies
func (d *DOM) buildRequest(method string, targetURL *url.URL) Request {
	req := Request{
		Method: method,
		URL:    targetURL,
		Header: make(http.Header),
		Form:   make(url.Values),
	}

	// Add cookies
	if len(d.cookieJar.cookies) > 0 {
		var cookieParts []string
		for name, value := range d.cookieJar.cookies {
			cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", name, value))
		}
		req.Header.Set("Cookie", strings.Join(cookieParts, "; "))
	}

	return req
}

// updateCookies updates cookies from the response
func (d *DOM) updateCookies(resp Response) {
	setCookieHeaders := resp.Header.Values("Set-Cookie")
	for _, header := range setCookieHeaders {
		// Simple cookie parsing (name=value; other attributes...)
		parts := strings.Split(header, ";")
		if len(parts) > 0 {
			cookiePair := strings.TrimSpace(parts[0])
			if idx := strings.Index(cookiePair, "="); idx > 0 {
				name := cookiePair[:idx]
				value := cookiePair[idx+1:]
				d.cookieJar.cookies[name] = value
			}
		}
	}
}

// Title returns the page title
func (d *DOM) Title() (string, error) {
	if d.doc == nil {
		return "", fmt.Errorf("document not parsed")
	}
	return d.doc.Find("title").Text(), nil
}

// Meta returns the content of a meta tag by name
func (d *DOM) Meta(name string) (string, error) {
	if d.doc == nil {
		return "", fmt.Errorf("document not parsed")
	}

	content, exists := d.doc.Find(fmt.Sprintf("meta[name=\"%s\"]", name)).Attr("content")
	if !exists {
		return "", fmt.Errorf("meta tag not found: %s", name)
	}

	return content, nil
}

// Text returns the text content of an element matching the selector
func (d *DOM) Text(selector string) (string, error) {
	sel, err := d.findOne(selector)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(sel.Text()), nil
}

// Find finds and returns a generic element matching the selector
// Supports standard CSS selectors
func (d *DOM) Find(selector string) (*Element, error) {
	if d.doc == nil {
		return nil, fmt.Errorf("document not parsed")
	}

	sel := d.doc.Find(selector)

	if sel.Length() == 0 {
		return nil, fmt.Errorf("element not found: %s", selector)
	}

	if sel.Length() > 1 {
		return nil, fmt.Errorf("multiple elements found for selector: %s (found %d)", selector, sel.Length())
	}

	return &Element{selection: sel}, nil
}

// Get makes a GET request to the specified URL and returns a new DOM with the response
func (d *DOM) Get(urlStr string) (*DOM, error) {
	targetURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Make relative URLs absolute
	if !targetURL.IsAbs() {
		targetURL = d.baseURL.ResolveReference(targetURL)
	}

	// Build GET request
	req := d.buildRequest(http.MethodGet, targetURL)

	// Send request
	resp, err := d.transport.Do(req)
	if err != nil {
		return nil, err
	}

	// Update cookies from response
	d.updateCookies(resp)

	// Create new DOM with same transport and cookie jar
	newDOM := &DOM{
		transport: d.transport,
		baseURL:   d.baseURL,
		cookieJar: d.cookieJar,
	}

	return newDOM.Parse(resp.Body), nil
}

// MustGet is like Get but panics on error
func (d *DOM) MustGet(urlStr string) *DOM {
	dom, err := d.Get(urlStr)
	if err != nil {
		panic(err)
	}
	return dom
}
