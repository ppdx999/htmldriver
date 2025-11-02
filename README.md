# htmldriver

[![Test](https://github.com/ppdx999/htmldriver/actions/workflows/test.yml/badge.svg)](https://github.com/ppdx999/htmldriver/actions/workflows/test.yml)
[![Lint](https://github.com/ppdx999/htmldriver/actions/workflows/lint.yml/badge.svg)](https://github.com/ppdx999/htmldriver/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ppdx999/htmldriver)](https://goreportcard.com/report/github.com/ppdx999/htmldriver)
[![codecov](https://codecov.io/gh/ppdx999/htmldriver/branch/main/graph/badge.svg)](https://codecov.io/gh/ppdx999/htmldriver)
[![GoDoc](https://godoc.org/github.com/ppdx999/htmldriver?status.svg)](https://godoc.org/github.com/ppdx999/htmldriver)

**A lightweight Go library for automated testing of pure HTML with human-like interactions.**

* Load existing HTML strings (SSR output, templates, files, snapshots)
* Emulate form input/submission, link/button clicks
* Verify headings, lists, tables, and arbitrary text
* **No JavaScript execution** (limited to pure HTML + Web standards)

[日本語ドキュメント (Japanese)](./README.ja.md)

---

## Table of Contents

* [Why htmldriver?](#why-htmldriver)
* [Features](#features)
* [Limitations](#limitations)
* [Installation](#installation)
* [Quick Start](#quick-start)
* [Form Operations](#form-operations)
* [Link Operations](#link-operations)
* [Finding Arbitrary Elements](#finding-arbitrary-elements)
* [Making GET Requests](#making-get-requests)
* [Table / List Verification](#table--list-verification)
* [Continuous Operations](#continuous-operations)
* [Cookie Management](#cookie-management)
* [Selectors](#selectors)
* [Transport](#transport)
* [API Overview](#api-overview)
* [Framework Integrations](#framework-integrations)
* [Roadmap](#roadmap)

---

## Why htmldriver?

When testing SSR applications or HTML templates in Go, existing tools have limitations:

* **goquery**: Specialized for DOM parsing, lacks form submission and HTTP request capabilities
* **net/http/httptest**: Can build mock servers, but form operation APIs aren't intuitive
* **playwright-go / chromedp**: Headless browsers are heavy, slow to start, and overkill when JavaScript execution is unnecessary

`htmldriver` is optimized for SSR/template testing with these characteristics:

* **Lightweight**: Only HTML parsing and HTTP requests (no browser required)
* **Intuitive API**: Human-like expressions such as `Fill()`, `Submit()`, `Click()`
* **Transport abstraction**: Integrates with any HTTP client/framework
* **Test framework agnostic**: No dependency on the `testing` package

## Features

* **Simple API**: Intuitive operations for form input and link clicks
* **Transport abstraction**: Works with any HTTP client/server
* **Automatic cookie management**: Supports session maintenance and authentication scenarios
* **Lightweight**: Few dependencies, easy setup

## Limitations

* **No JavaScript support**: Limited to pure HTML + Web standard behavior
* **No CSS/layout evaluation**: Cannot detect visual breakage (text/DOM structure only)
* **No browser-specific behavior**: Does not reproduce browser implementation differences

If you need JavaScript execution, consider headless browser libraries like `playwright-go` or `chromedp`.

## Installation

```bash
go get github.com/ppdx999/htmldriver
```

## Quick Start

```go
package login_test

import (
    "net/http"
    "testing"

    h "github.com/ppdx999/htmldriver"
)

type MockTransport struct{}

func (m MockTransport) Do(req h.Request) (h.Response, error) {
    // Return arbitrary response based on submitted form/URL
    if req.Method == http.MethodPost && req.URL.Path == "/login" {
        user := req.Form.Get("username")
        return h.Response{Status: 200, Body: "<p>Welcome, " + user + "</p>"}, nil
    }
    return h.Response{Status: 404, Body: "not found"}, nil
}

func Test_LoginFlow(t *testing.T) {
    html := `
    <form id="login-form" action="/login" method="post">
      <label>User</label><input type="text" name="username">
      <label>Pass</label><input type="password" name="password">
      <button type="submit">Login</button>
    </form>`

    dom := h.New(MockTransport{}).Parse(html)

    form, err := dom.Form("#login-form")
    if err != nil {
        t.Fatal(err)
    }

    form.MustFill("username", "alice").MustFill("password", "secret")

    res, err := form.Submit()
    if err != nil {
        t.Fatal(err)
    }

    if res.Status != 200 {
        t.Fatalf("expected status 200, got %d", res.Status)
    }
}
```

## Form Operations

### Error Handling Version

```go
form, err := dom.Form("@login")
if err != nil {
    return err
}

form, err = form.Fill("username", "tester")
if err != nil {
    return err
}

form, err = form.Fill("password", "mypassword")
if err != nil {
    return err
}

form, err = form.CheckCheckbox("remember_me")
if err != nil {
    return err
}

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

### Method Chaining Version (Must*)

For tests where you want to fail immediately on error, use `Must` prefixed methods. These panic on error.

```go
form, err := dom.Form("@login")
if err != nil {
    return err
}

form.MustFill("username", "tester").
    MustFill("password", "mypassword").
    MustCheckCheckbox("remember_me")

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

## Link Operations

```go
link, err := dom.Link("@profile")
if err != nil {
    return err
}

text, err := link.GetText()
if err != nil {
    return err
}
if text != "View Profile" {
    return fmt.Errorf("expected link text 'View Profile', got '%s'", text)
}

url, err := link.GetURL()
if err != nil {
    return err
}
if url.Path != "/users/123" {
    return fmt.Errorf("expected link URL '/users/123', got '%s'", url.Path)
}

res, err := link.Click()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

## Finding Arbitrary Elements

You can find any HTML element using standard CSS selectors with `Find()`:

```go
// Find element by class
errorAlert, err := dom.Find(".alert-danger")
if err != nil {
    return err
}

// Get element text
errorText := errorAlert.Text()

// Get element attributes
className := errorAlert.Attr("class")
dataValue := errorAlert.Attr("data-value")

// Check if attribute exists
if errorAlert.HasAttr("data-error") {
    // Handle error attribute
}

// Get HTML content
html, err := errorAlert.HTML()

// Check element type
if errorAlert.Is(".alert") {
    // It's an alert
}
```

You can also find elements within forms:

```go
form, _ := dom.Form("#login-form")

// Find input field within form
emailField, err := form.Find("#email")
if err != nil {
    return err
}

// Get field value from attribute
emailValue := emailField.Attr("value")
```

## Making GET Requests

You can make GET requests directly and get a new DOM:

```go
// Create initial DOM
dom := h.New(transport).Parse(html)

// Make GET request and get new DOM
homePage, err := dom.Get("/home")
if err != nil {
    return err
}

// Or use MustGet (panics on error)
homePage := dom.MustGet("/home")

// Cookies are automatically carried forward
profilePage := homePage.MustGet("/profile")
```

## Table / List Verification

### Table

```go
table, err := dom.Table("@user-list")
if err != nil {
    return err
}

rows, err := table.GetRows()
if err != nil {
    return err
}

// Verify row count
if table.GetRowCount() != 3 {
    return fmt.Errorf("expected 3 rows, got %d", table.GetRowCount())
}

// Verify cell content
if rows[0][0] != "Alice" {
    return fmt.Errorf("expected 'Alice', got '%s'", rows[0][0])
}

// Check for empty table
if table.GetRowCount() == 0 {
    return fmt.Errorf("expected data, but table is empty")
}
```

### List

```go
list, err := dom.List("@todo-items")
if err != nil {
    return err
}

items, err := list.GetItems()
if err != nil {
    return err
}

// Verify item count
if list.GetItemCount() != 5 {
    return fmt.Errorf("expected 5 items, got %d", list.GetItemCount())
}

// Verify content
if items[0] != "Buy groceries" {
    return fmt.Errorf("expected 'Buy groceries', got '%s'", items[0])
}
```

**Note**: Advanced comparison and diff display are planned in the roadmap.

## Continuous Operations

`Submit()` and `Click()` return a `Response` containing HTML body, which can be parsed again for the next operation.

### Method 1: Share CookieJar (Recommended)

To share cookies across multiple page operations, explicitly create and share a `CookieJar`.

```go
// Create transport
transport := NewMockTransport()

// Create CookieJar for sharing cookies
jar := h.NewCookieJar()

// Login page
dom := h.NewWithCookieJar(transport, jar).Parse(loginPageHTML)

form, err := dom.Form("#login-form")
if err != nil {
    return err
}

form.MustFill("username", "alice").MustFill("password", "secret")

res, err := form.Submit()
if err != nil {
    return err
}

// Parse dashboard (use same jar)
dashboardDOM := h.NewWithCookieJar(transport, jar).Parse(res.Body)

// Click edit profile link
link, err := dashboardDOM.Link("@edit-profile")
if err != nil {
    return err
}

res, err = link.Click()
if err != nil {
    return err
}

// Parse edit page (use same jar)
editDOM := h.NewWithCookieJar(transport, jar).Parse(res.Body)

form, err = editDOM.Form("#profile-form")
if err != nil {
    return err
}

form.MustFill("bio", "Updated bio text")

res, err = form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

### Method 2: Reuse the Same DOM Instance

To automatically carry cookies forward, you can also repeatedly call `Parse()` on the same DOM instance.

```go
transport := NewMockTransport()

// First page
dom := h.New(transport).Parse(loginPageHTML)

// Login
form, _ := dom.Form("#login-form")
form.MustFill("username", "alice").MustFill("password", "secret")
res, _ := form.Submit()

// Parse next page with same DOM (cookies are automatically carried)
dom = dom.Parse(res.Body)

// Continue using dom...
```

**Important**: Creating a new DOM with `New()` does not carry cookies forward. Use `NewWithCookieJar()` to share a `CookieJar`, or keep using the same DOM instance.

## Cookie Management

`htmldriver` automatically saves cookies from responses and sends them with subsequent form submissions or link clicks. This makes it easy to test session management and user authentication scenarios.

To set cookies beforehand:

```go
dom := h.New(transport).Parse(html)
dom.SetCookie("session_id", "abc123")

// Subsequent operations will send "session_id=abc123"
form, err := dom.Form("#protected-form")
if err != nil {
    return err
}

form.MustFill("data", "value")

res, err := form.Submit()
if err != nil {
    return err
}

if res.Status != 200 {
    return fmt.Errorf("expected status 200, got %d", res.Status)
}
```

**Note**: Detailed cookie attributes (expiration, path, domain, Secure, HttpOnly, etc.) are planned in the roadmap.

## Selectors

The `selector` parameter accepts two types of locator strings:

| Selector | Description |
|----------|-------------|
| `@xxxxx` | Matches elements with `test-id` attribute equal to `xxxxx` |
| `#xxxxx` | Matches elements with `id` attribute equal to `xxxxx` |

Support for more advanced CSS selectors is being considered in the roadmap.

## Transport

`Transport` is an interface that abstracts I/O operations.

`Form.Submit()` and `Link.Click()` internally call `Transport.Do()` to send HTTP requests and return the results to the caller.

This mechanism allows `htmldriver` to work with any HTTP client or server framework without being tied to a specific implementation.

```go
type Request struct {
    Method string
    URL    *url.URL
    Header http.Header
    Form   url.Values    // for x-www-form-urlencoded
    Files  []FormFile    // for multipart
}

type Response struct {
    Header http.Header
    URL    *url.URL
    Body   string
    Status int
}

type Transport interface {
    Do(req Request) (Response, error)
}
```

### Example: Using Standard http.Client

```go
type HTTPClientTransport struct {
    client  *http.Client
    baseURL string
}

func NewHTTPClientTransport(baseURL string) *HTTPClientTransport {
    return &HTTPClientTransport{
        client:  &http.Client{},
        baseURL: baseURL,
    }
}

func (t *HTTPClientTransport) Do(req h.Request) (h.Response, error) {
    // Convert relative URL to absolute URL
    fullURL := t.baseURL + req.URL.String()

    var httpReq *http.Request
    var err error

    if req.Method == http.MethodPost {
        httpReq, err = http.NewRequest(req.Method, fullURL, strings.NewReader(req.Form.Encode()))
        if err != nil {
            return h.Response{}, err
        }
        httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    } else {
        httpReq, err = http.NewRequest(req.Method, fullURL, nil)
        if err != nil {
            return h.Response{}, err
        }
    }

    // Copy headers
    for k, v := range req.Header {
        httpReq.Header[k] = v
    }

    resp, err := t.client.Do(httpReq)
    if err != nil {
        return h.Response{}, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return h.Response{}, err
    }

    return h.Response{
        Header: resp.Header,
        URL:    resp.Request.URL,
        Body:   string(body),
        Status: resp.StatusCode,
    }, nil
}
```

### Base URL Handling

When form `action` or `<a>` `href` are relative paths, the Transport manages the base URL.

- **Default**: Uses `http://localhost`
- **Customization**: Specify `baseURL` when implementing Transport

## API Overview

```go
// Root
func New(transport Transport) *DOM
func NewWithCookieJar(transport Transport, jar *CookieJar) *DOM
func NewCookieJar() *CookieJar
func (d *DOM) Parse(html string) *DOM
func (d *DOM) SetCookie(name, value string) *DOM
func (d *DOM) GetCookieJar() *CookieJar
func (d *DOM) Get(url string) (*DOM, error)                    // Make GET request
func (d *DOM) MustGet(url string) *DOM                         // Make GET request (panic on error)

// Element Retrieval
func (d *DOM) Form(selector string) (*Form, error)
func (d *DOM) Link(selector string) (*Link, error)
func (d *DOM) Button(selector string) (*Button, error)
func (d *DOM) Table(selector string) (*Table, error)
func (d *DOM) List(selector string) (*List, error)
func (d *DOM) Text(selector string) (string, error)
func (d *DOM) Title() (string, error)
func (d *DOM) Meta(name string) (string, error)
func (d *DOM) Img(selector string) (*Img, error)
func (d *DOM) Find(selector string) (*Element, error)         // Find arbitrary element

// Form Operations (Error-returning version)
func (f *Form) Fill(name, value string) (*Form, error)           // Text input
func (f *Form) CheckCheckbox(name string) (*Form, error)         // Check checkbox
func (f *Form) UncheckCheckbox(name string) (*Form, error)       // Uncheck checkbox
func (f *Form) Select(name, value string) (*Form, error)         // Select option
func (f *Form) CheckRadio(name, value string) (*Form, error)     // Select radio
func (f *Form) Submit() (Response, error)                        // Submit

// Form Operations (Method chaining version - panics on error)
func (f *Form) MustFill(name, value string) *Form                // Text input
func (f *Form) MustCheckCheckbox(name string) *Form              // Check checkbox
func (f *Form) MustUncheckCheckbox(name string) *Form            // Uncheck checkbox
func (f *Form) MustSelect(name, value string) *Form              // Select option
func (f *Form) MustCheckRadio(name, value string) *Form          // Select radio

// Form Information
func (f *Form) GetValue(name string) (string, error)             // Get value
func (f *Form) HasField(name string) bool                        // Check field exists
func (f *Form) Find(selector string) (*Element, error)           // Find element in form

// Element
func (e *Element) Text() string                                  // Get text content
func (e *Element) Attr(name string) string                       // Get attribute value
func (e *Element) HasAttr(name string) bool                      // Check if attribute exists
func (e *Element) HTML() (string, error)                         // Get inner HTML
func (e *Element) Is(selector string) bool                       // Check if matches selector

// Link
func (l *Link) Click() (Response, error)
func (l *Link) GetURL() (*url.URL, error)
func (l *Link) GetText() (string, error)

// Button
func (b *Button) GetText() (string, error)

// Table
func (tbl *Table) GetRows() ([][]string, error)
func (tbl *Table) GetRowCount() int
func (tbl *Table) GetColCount() int

// List
func (lst *List) GetItems() ([]string, error)
func (lst *List) GetItemCount() int

// Image
func (img *Img) GetSrc() (string, error)
func (img *Img) GetAlt() (string, error)
```

## Framework Integrations

Transport implementations for major HTTP server frameworks are provided.

### Chi Integration

When using the Chi framework, you can use `ChiTransport` to work with `httptest.Server`.

```go
import (
    "github.com/go-chi/chi/v5"
    h "github.com/ppdx999/htmldriver"
    "github.com/ppdx999/htmldriver/integrations/chitransport"
)

func Test_LoginFlow(t *testing.T) {
    r := chi.NewRouter()

    transport := chitransport.NewChiTransport(r)

    dom := h.New(transport).Parse(renderLoginPage())

    form, err := dom.Form("#login-form")
    if err != nil {
        t.Fatal(err)
    }

    form.MustFill("username", "alice").
        MustFill("password", "secret")

    res, err := form.Submit()
    if err != nil {
        t.Fatal(err)
    }

    if res.Status != 200 {
        t.Fatalf("expected status 200, got %d", res.Status)
    }
}
```

### Echo Integration

Coming soon...

### Gin Integration

Coming soon...

### Fiber Integration

Coming soon...

## Roadmap

* [ ] `multipart/form-data` and file upload support
* [ ] `<select multiple>` / `<input type=date|time|number>` input helpers
* [ ] Redirect following (`3xx`) support
* [ ] Detailed cookie attribute support (expiration, path, domain, Secure, HttpOnly, etc.)
* [ ] Enhanced `Table`/`List` diff reporting (detailed cell mismatch display)
* [ ] Rich selector extensions (class, tag, attribute selectors, `:has()`, `:nth-of-type()`, etc.)
* [ ] Improved error messages with color diff output

## License

MIT License - see [LICENSE](LICENSE) file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
