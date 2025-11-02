package htmldriver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// Test Element.Text()
func TestElement_Text(t *testing.T) {
	html := `<div class="message">  Hello, World!  </div>`

	dom := New(NewMockTransport()).Parse(html)

	elem, err := dom.Find(".message")
	if err != nil {
		t.Fatal(err)
	}

	text := elem.Text()
	if text != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got '%s'", text)
	}
}

// Test Element.Attr()
func TestElement_Attr(t *testing.T) {
	html := `<input id="email" type="email" value="test@example.com" placeholder="Enter email">`

	dom := New(NewMockTransport()).Parse(html)

	elem, err := dom.Find("#email")
	if err != nil {
		t.Fatal(err)
	}

	// Test existing attribute
	emailType := elem.Attr("type")
	if emailType != "email" {
		t.Fatalf("expected 'email', got '%s'", emailType)
	}

	value := elem.Attr("value")
	if value != "test@example.com" {
		t.Fatalf("expected 'test@example.com', got '%s'", value)
	}

	// Test non-existent attribute
	nonExistent := elem.Attr("nonexistent")
	if nonExistent != "" {
		t.Fatalf("expected empty string for non-existent attribute, got '%s'", nonExistent)
	}
}

// Test Element.HasAttr()
func TestElement_HasAttr(t *testing.T) {
	html := `<input type="checkbox" checked>`

	dom := New(NewMockTransport()).Parse(html)

	elem, err := dom.Find("input")
	if err != nil {
		t.Fatal(err)
	}

	if !elem.HasAttr("type") {
		t.Fatal("expected element to have 'type' attribute")
	}

	if !elem.HasAttr("checked") {
		t.Fatal("expected element to have 'checked' attribute")
	}

	if elem.HasAttr("nonexistent") {
		t.Fatal("expected element to not have 'nonexistent' attribute")
	}
}

// Test Element.HTML()
func TestElement_HTML(t *testing.T) {
	html := `<div class="container"><span>Content</span></div>`

	dom := New(NewMockTransport()).Parse(html)

	elem, err := dom.Find(".container")
	if err != nil {
		t.Fatal(err)
	}

	innerHTML, err := elem.HTML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(innerHTML, "<span>Content</span>") {
		t.Fatalf("expected HTML to contain '<span>Content</span>', got '%s'", innerHTML)
	}
}

// Test Element.Is()
func TestElement_Is(t *testing.T) {
	html := `<button class="btn btn-primary" type="submit">Submit</button>`

	dom := New(NewMockTransport()).Parse(html)

	elem, err := dom.Find("button")
	if err != nil {
		t.Fatal(err)
	}

	if !elem.Is("button") {
		t.Fatal("expected element to match 'button' selector")
	}

	if !elem.Is(".btn") {
		t.Fatal("expected element to match '.btn' selector")
	}

	if !elem.Is("[type='submit']") {
		t.Fatal("expected element to match '[type='submit']' selector")
	}

	if elem.Is("a") {
		t.Fatal("expected element to not match 'a' selector")
	}
}

// Test DOM.Find() with CSS selectors
func TestDOM_Find(t *testing.T) {
	html := `
	<div class="alert alert-danger">Error occurred</div>
	<p id="description">Description text</p>
	<span data-value="123">Value</span>
	`

	dom := New(NewMockTransport()).Parse(html)

	// Test class selector
	alert, err := dom.Find(".alert-danger")
	if err != nil {
		t.Fatal(err)
	}
	if alert.Text() != "Error occurred" {
		t.Fatalf("expected 'Error occurred', got '%s'", alert.Text())
	}

	// Test ID selector
	desc, err := dom.Find("#description")
	if err != nil {
		t.Fatal(err)
	}
	if desc.Text() != "Description text" {
		t.Fatalf("expected 'Description text', got '%s'", desc.Text())
	}

	// Test attribute selector
	span, err := dom.Find("[data-value='123']")
	if err != nil {
		t.Fatal(err)
	}
	if span.Attr("data-value") != "123" {
		t.Fatalf("expected '123', got '%s'", span.Attr("data-value"))
	}
}

// Test DOM.Find() error cases
func TestDOM_Find_Errors(t *testing.T) {
	html := `<div class="item">Item 1</div><div class="item">Item 2</div>`

	dom := New(NewMockTransport()).Parse(html)

	// Test not found
	_, err := dom.Find(".nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent element")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %s", err.Error())
	}

	// Test multiple elements
	_, err = dom.Find(".item")
	if err == nil {
		t.Fatal("expected error for multiple elements")
	}
	if !strings.Contains(err.Error(), "multiple elements") {
		t.Fatalf("expected 'multiple elements' error, got: %s", err.Error())
	}
}

// Test Form.Find()
func TestForm_Find(t *testing.T) {
	html := `
	<form id="login-form">
		<input id="email" type="email" name="email">
		<input id="password" type="password" name="password">
		<button type="submit">Login</button>
		<span class="error-message">Invalid credentials</span>
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#login-form")
	if err != nil {
		t.Fatal(err)
	}

	// Find input field
	emailField, err := form.Find("#email")
	if err != nil {
		t.Fatal(err)
	}
	if emailField.Attr("type") != "email" {
		t.Fatalf("expected 'email', got '%s'", emailField.Attr("type"))
	}

	// Find error message
	errorMsg, err := form.Find(".error-message")
	if err != nil {
		t.Fatal(err)
	}
	if errorMsg.Text() != "Invalid credentials" {
		t.Fatalf("expected 'Invalid credentials', got '%s'", errorMsg.Text())
	}

	// Find button
	button, err := form.Find("button[type='submit']")
	if err != nil {
		t.Fatal(err)
	}
	if button.Text() != "Login" {
		t.Fatalf("expected 'Login', got '%s'", button.Text())
	}
}

// Test DOM.Get() and MustGet()
func TestDOM_Get(t *testing.T) {
	transport := NewMockTransport()
	transport.AddResponse("GET", "/page", 200, `<h1>Page Content</h1><p>Text</p>`)

	dom := New(transport).Parse("<html></html>")

	// Test Get()
	newDom, err := dom.Get("/page")
	if err != nil {
		t.Fatal(err)
	}

	h1, err := newDom.Find("h1")
	if err != nil {
		t.Fatal(err)
	}
	if h1.Text() != "Page Content" {
		t.Fatalf("expected 'Page Content', got '%s'", h1.Text())
	}

	// Test MustGet()
	newDom2 := dom.MustGet("/page")
	p, err := newDom2.Find("p")
	if err != nil {
		t.Fatal(err)
	}
	if p.Text() != "Text" {
		t.Fatalf("expected 'Text', got '%s'", p.Text())
	}
}

// Test DOM.Get() with cookie persistence
func TestDOM_Get_CookiePersistence(t *testing.T) {
	transport := NewMockTransport()

	// First response sets a cookie
	resp1 := Response{
		Status: 200,
		Body:   `<a href="/protected">Go to protected</a>`,
		Header: http.Header{},
		URL:    &url.URL{Path: "/login"},
	}
	resp1.Header.Add("Set-Cookie", "session=xyz; Path=/")
	transport.responses["GET /login"] = resp1

	// Second response
	transport.AddResponse("GET", "/protected", 200, `<h1>Protected Content</h1>`)

	jar := NewCookieJar()
	dom := NewWithCookieJar(transport, jar).Parse("<html></html>")

	// First request
	dom1, err := dom.Get("/login")
	if err != nil {
		t.Fatal(err)
	}

	// Check cookie was saved
	if jar.cookies["session"] != "xyz" {
		t.Fatalf("expected session cookie 'xyz', got '%s'", jar.cookies["session"])
	}

	// Second request should have the cookie
	dom2, err := dom1.Get("/protected")
	if err != nil {
		t.Fatal(err)
	}

	h1, err := dom2.Find("h1")
	if err != nil {
		t.Fatal(err)
	}
	if h1.Text() != "Protected Content" {
		t.Fatalf("expected 'Protected Content', got '%s'", h1.Text())
	}
}

// ErrorTransport for testing error cases
type ErrorTransport struct{}

func (e ErrorTransport) Do(req Request) (Response, error) {
	return Response{}, fmt.Errorf("transport error")
}

// Test MustGet panic
func TestDOM_MustGet_Panic(t *testing.T) {
	// Create transport that returns an error
	errorTransport := ErrorTransport{}

	dom := New(errorTransport).Parse("<html></html>")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when using MustGet with transport error")
		}
	}()

	dom.MustGet("/any")
}

// Example usage of new APIs
func ExampleElement() {
	html := `
	<div class="alert alert-danger">
		<strong>Error:</strong> Invalid input
	</div>`

	dom := New(NewMockTransport()).Parse(html)

	// Find element by CSS selector
	alert, _ := dom.Find(".alert-danger")

	// Get text content
	text := alert.Text()
	_ = text // "Error: Invalid input"

	// Get attribute
	className := alert.Attr("class")
	_ = className // "alert alert-danger"
}

func ExampleDOM_Get() {
	transport := NewMockTransport()
	transport.AddResponse("GET", "/home", 200, "<h1>Home</h1>")

	dom := New(transport).Parse("<html></html>")

	// Make GET request and get new DOM
	homeDom := dom.MustGet("/home")

	h1, _ := homeDom.Find("h1")
	_ = h1.Text() // "Home"
}
