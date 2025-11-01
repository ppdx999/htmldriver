package htmldriver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// MockTransport is a simple transport for testing
type MockTransport struct {
	responses map[string]Response
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		responses: make(map[string]Response),
	}
}

func (m *MockTransport) AddResponse(method, path string, status int, body string) {
	key := method + " " + path
	m.responses[key] = Response{
		Status: status,
		Body:   body,
		Header: make(http.Header),
		URL:    &url.URL{Path: path},
	}
}

func (m *MockTransport) Do(req Request) (Response, error) {
	key := req.Method + " " + req.URL.Path
	resp, ok := m.responses[key]
	if !ok {
		return Response{
			Status: 404,
			Body:   "not found",
			Header: make(http.Header),
			URL:    req.URL,
		}, nil
	}
	return resp, nil
}

// Test basic form submission from README example
func TestLoginFlow(t *testing.T) {
	transport := NewMockTransport()
	transport.AddResponse("POST", "/login", 200, "<p>Welcome, alice</p>")

	html := `
	<form id="login-form" action="/login" method="post">
		<label>User</label><input type="text" name="username">
		<label>Pass</label><input type="password" name="password">
		<button type="submit">Login</button>
	</form>`

	dom := New(transport).Parse(html)

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

	if !strings.Contains(res.Body, "Welcome, alice") {
		t.Fatalf("expected 'Welcome, alice' in response, got: %s", res.Body)
	}
}

// Test selector resolution
func TestSelectorResolution(t *testing.T) {
	html := `
	<div id="by-id">ID Selector</div>
	<div test-id="by-test-id">Test-ID Selector</div>
	`

	dom := New(NewMockTransport()).Parse(html)

	// Test ID selector
	text, err := dom.Text("#by-id")
	if err != nil {
		t.Fatal(err)
	}
	if text != "ID Selector" {
		t.Fatalf("expected 'ID Selector', got '%s'", text)
	}

	// Test test-id selector
	text, err = dom.Text("@by-test-id")
	if err != nil {
		t.Fatal(err)
	}
	if text != "Test-ID Selector" {
		t.Fatalf("expected 'Test-ID Selector', got '%s'", text)
	}
}

// Test form checkbox operations
func TestFormCheckbox(t *testing.T) {
	html := `
	<form id="settings-form">
		<input type="checkbox" name="remember_me" value="yes">
		<input type="checkbox" name="newsletter" value="yes" checked>
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#settings-form")
	if err != nil {
		t.Fatal(err)
	}

	// Check that "newsletter" is checked by default
	value, err := form.GetValue("newsletter")
	if err != nil {
		t.Fatal(err)
	}
	if value != "yes" {
		t.Fatalf("expected 'yes', got '%s'", value)
	}

	// Check "remember_me"
	form, err = form.CheckCheckbox("remember_me")
	if err != nil {
		t.Fatal(err)
	}

	value, err = form.GetValue("remember_me")
	if err != nil {
		t.Fatal(err)
	}
	if value != "yes" {
		t.Fatalf("expected 'yes', got '%s'", value)
	}

	// Uncheck "newsletter"
	form, err = form.UncheckCheckbox("newsletter")
	if err != nil {
		t.Fatal(err)
	}

	value, _ = form.GetValue("newsletter")
	if value != "" {
		t.Fatalf("expected empty string, got '%s'", value)
	}
}

// Test form radio buttons
func TestFormRadio(t *testing.T) {
	html := `
	<form id="survey-form">
		<input type="radio" name="color" value="red">
		<input type="radio" name="color" value="blue">
		<input type="radio" name="color" value="green">
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#survey-form")
	if err != nil {
		t.Fatal(err)
	}

	// Select blue
	form, err = form.CheckRadio("color", "blue")
	if err != nil {
		t.Fatal(err)
	}

	value, err := form.GetValue("color")
	if err != nil {
		t.Fatal(err)
	}
	if value != "blue" {
		t.Fatalf("expected 'blue', got '%s'", value)
	}
}

// Test form select element
func TestFormSelect(t *testing.T) {
	html := `
	<form id="country-form">
		<select name="country">
			<option value="us">United States</option>
			<option value="jp" selected>Japan</option>
			<option value="uk">United Kingdom</option>
		</select>
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#country-form")
	if err != nil {
		t.Fatal(err)
	}

	// Check default value
	value, err := form.GetValue("country")
	if err != nil {
		t.Fatal(err)
	}
	if value != "jp" {
		t.Fatalf("expected 'jp', got '%s'", value)
	}

	// Change to UK
	form, err = form.Select("country", "uk")
	if err != nil {
		t.Fatal(err)
	}

	value, err = form.GetValue("country")
	if err != nil {
		t.Fatal(err)
	}
	if value != "uk" {
		t.Fatalf("expected 'uk', got '%s'", value)
	}
}

// Test link operations
func TestLink(t *testing.T) {
	transport := NewMockTransport()
	transport.AddResponse("GET", "/profile", 200, "<h1>Profile Page</h1>")

	html := `<a id="profile-link" href="/profile">View Profile</a>`

	dom := New(transport).Parse(html)

	link, err := dom.Link("#profile-link")
	if err != nil {
		t.Fatal(err)
	}

	text, err := link.GetText()
	if err != nil {
		t.Fatal(err)
	}
	if text != "View Profile" {
		t.Fatalf("expected 'View Profile', got '%s'", text)
	}

	linkURL, err := link.GetURL()
	if err != nil {
		t.Fatal(err)
	}
	if linkURL.Path != "/profile" {
		t.Fatalf("expected '/profile', got '%s'", linkURL.Path)
	}

	res, err := link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}
}

// Test table operations
func TestTable(t *testing.T) {
	html := `
	<table id="user-list">
		<tr>
			<th>Name</th>
			<th>Age</th>
		</tr>
		<tr>
			<td>Alice</td>
			<td>30</td>
		</tr>
		<tr>
			<td>Bob</td>
			<td>25</td>
		</tr>
	</table>`

	dom := New(NewMockTransport()).Parse(html)

	table, err := dom.Table("#user-list")
	if err != nil {
		t.Fatal(err)
	}

	rows, err := table.GetRows()
	if err != nil {
		t.Fatal(err)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[0][0] != "Name" || rows[0][1] != "Age" {
		t.Fatalf("unexpected header row: %v", rows[0])
	}

	if rows[1][0] != "Alice" || rows[1][1] != "30" {
		t.Fatalf("unexpected data row: %v", rows[1])
	}

	if table.GetRowCount() != 3 {
		t.Fatalf("expected row count 3, got %d", table.GetRowCount())
	}

	if table.GetColCount() != 2 {
		t.Fatalf("expected col count 2, got %d", table.GetColCount())
	}
}

// Test list operations
func TestList(t *testing.T) {
	html := `
	<ul id="todo-items">
		<li>Buy groceries</li>
		<li>Walk the dog</li>
		<li>Write code</li>
	</ul>`

	dom := New(NewMockTransport()).Parse(html)

	list, err := dom.List("#todo-items")
	if err != nil {
		t.Fatal(err)
	}

	items, err := list.GetItems()
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	if items[0] != "Buy groceries" {
		t.Fatalf("expected 'Buy groceries', got '%s'", items[0])
	}

	if list.GetItemCount() != 3 {
		t.Fatalf("expected item count 3, got %d", list.GetItemCount())
	}
}

// Test cookie handling
func TestCookies(t *testing.T) {
	transport := NewMockTransport()

	// Setup a response that sets cookies
	resp := Response{
		Status: 200,
		Body:   "<p>Logged in</p>",
		Header: http.Header{},
		URL:    &url.URL{Path: "/login"},
	}
	resp.Header.Add("Set-Cookie", "session_id=abc123; Path=/")
	resp.Header.Add("Set-Cookie", "user=alice; Path=/")

	transport.responses["POST /login"] = resp

	html := `
	<form id="login-form" action="/login" method="post">
		<input type="text" name="username">
		<input type="password" name="password">
	</form>`

	dom := New(transport).Parse(html)

	form, err := dom.Form("#login-form")
	if err != nil {
		t.Fatal(err)
	}

	form.MustFill("username", "alice").MustFill("password", "secret")

	_, err = form.Submit()
	if err != nil {
		t.Fatal(err)
	}

	// Check that cookies were saved
	jar := dom.GetCookieJar()
	if jar.cookies["session_id"] != "abc123" {
		t.Fatalf("expected session_id cookie 'abc123', got '%s'", jar.cookies["session_id"])
	}

	if jar.cookies["user"] != "alice" {
		t.Fatalf("expected user cookie 'alice', got '%s'", jar.cookies["user"])
	}
}

// Test error handling for Fill with wrong input type
func TestFormFillErrorHandling(t *testing.T) {
	html := `
	<form id="test-form">
		<input type="checkbox" name="accept" value="yes">
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#test-form")
	if err != nil {
		t.Fatal(err)
	}

	// Try to use Fill on a checkbox (should error)
	_, err = form.Fill("accept", "yes")
	if err == nil {
		t.Fatal("expected error when using Fill on checkbox")
	}

	if !strings.Contains(err.Error(), "checkbox") {
		t.Fatalf("expected error message to mention checkbox, got: %s", err.Error())
	}
}

// Test Must* methods panic on error
func TestMustMethodsPanic(t *testing.T) {
	html := `<form id="test-form"></form>`

	dom := New(NewMockTransport()).Parse(html)

	form, err := dom.Form("#test-form")
	if err != nil {
		t.Fatal(err)
	}

	// Should panic when field doesn't exist
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when using MustFill on non-existent field")
		}
	}()

	form.MustFill("nonexistent", "value")
}

// Test Title and Meta
func TestTitleAndMeta(t *testing.T) {
	html := `
	<html>
		<head>
			<title>Test Page</title>
			<meta name="description" content="This is a test page">
		</head>
		<body></body>
	</html>`

	dom := New(NewMockTransport()).Parse(html)

	title, err := dom.Title()
	if err != nil {
		t.Fatal(err)
	}
	if title != "Test Page" {
		t.Fatalf("expected 'Test Page', got '%s'", title)
	}

	meta, err := dom.Meta("description")
	if err != nil {
		t.Fatal(err)
	}
	if meta != "This is a test page" {
		t.Fatalf("expected 'This is a test page', got '%s'", meta)
	}
}

// Test Button
func TestButton(t *testing.T) {
	html := `
	<button id="submit-btn">Submit Form</button>
	<input id="input-btn" type="submit" value="Send">
	`

	dom := New(NewMockTransport()).Parse(html)

	// Test button element
	btn, err := dom.Button("#submit-btn")
	if err != nil {
		t.Fatal(err)
	}

	text, err := btn.GetText()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Submit Form" {
		t.Fatalf("expected 'Submit Form', got '%s'", text)
	}

	// Test input[type=submit]
	btn, err = dom.Button("#input-btn")
	if err != nil {
		t.Fatal(err)
	}

	text, err = btn.GetText()
	if err != nil {
		t.Fatal(err)
	}
	if text != "Send" {
		t.Fatalf("expected 'Send', got '%s'", text)
	}
}

// Test Img
func TestImg(t *testing.T) {
	html := `<img id="logo" src="/images/logo.png" alt="Company Logo">`

	dom := New(NewMockTransport()).Parse(html)

	img, err := dom.Img("#logo")
	if err != nil {
		t.Fatal(err)
	}

	src, err := img.GetSrc()
	if err != nil {
		t.Fatal(err)
	}
	if src != "/images/logo.png" {
		t.Fatalf("expected '/images/logo.png', got '%s'", src)
	}

	alt, err := img.GetAlt()
	if err != nil {
		t.Fatal(err)
	}
	if alt != "Company Logo" {
		t.Fatalf("expected 'Company Logo', got '%s'", alt)
	}
}

// Test continuous operations (login -> profile -> edit)
func TestContinuousOperations(t *testing.T) {
	transport := NewMockTransport()

	// Login response with cookie
	loginResp := Response{
		Status: 200,
		Body: `
			<h1>Dashboard</h1>
			<a test-id="edit-profile" href="/profile/edit">Edit Profile</a>
		`,
		Header: http.Header{},
		URL:    &url.URL{Path: "/login"},
	}
	loginResp.Header.Add("Set-Cookie", "session=xyz789; Path=/")
	transport.responses["POST /login"] = loginResp

	// Profile edit page
	transport.AddResponse("GET", "/profile/edit", 200, `
		<h1>Edit Profile</h1>
		<form id="profile-form" action="/profile/update" method="post">
			<input type="text" name="bio" value="Old bio">
		</form>
	`)

	// Profile update response
	transport.AddResponse("POST", "/profile/update", 200, "<p>Profile updated</p>")

	// Start with login page
	loginHTML := `
	<form id="login-form" action="/login" method="post">
		<input type="text" name="username">
		<input type="password" name="password">
	</form>`

	// Create shared cookie jar
	jar := NewCookieJar()
	dom := NewWithCookieJar(transport, jar).Parse(loginHTML)

	// Step 1: Login
	form, err := dom.Form("#login-form")
	if err != nil {
		t.Fatal(err)
	}

	form.MustFill("username", "alice").MustFill("password", "secret")

	res, err := form.Submit()
	if err != nil {
		t.Fatal(err)
	}

	// Step 2: Parse dashboard and click edit link (share cookie jar)
	dashboardDOM := NewWithCookieJar(transport, jar).Parse(res.Body)

	link, err := dashboardDOM.Link("@edit-profile")
	if err != nil {
		t.Fatal(err)
	}

	res, err = link.Click()
	if err != nil {
		t.Fatal(err)
	}

	// Step 3: Parse edit page and submit form (share cookie jar)
	editDOM := NewWithCookieJar(transport, jar).Parse(res.Body)

	form, err = editDOM.Form("#profile-form")
	if err != nil {
		t.Fatal(err)
	}

	// Check current value
	currentBio, err := form.GetValue("bio")
	if err != nil {
		t.Fatal(err)
	}
	if currentBio != "Old bio" {
		t.Fatalf("expected 'Old bio', got '%s'", currentBio)
	}

	form.MustFill("bio", "Updated bio text")

	res, err = form.Submit()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	if !strings.Contains(res.Body, "Profile updated") {
		t.Fatalf("expected 'Profile updated' in response")
	}

	// Verify that cookies persisted across operations
	if jar.cookies["session"] != "xyz789" {
		t.Fatal("expected session cookie to persist")
	}
}

// Test SetCookie
func TestSetCookie(t *testing.T) {
	html := `<form id="test-form" action="/test" method="post"></form>`

	transport := NewMockTransport()
	transport.AddResponse("POST", "/test", 200, "OK")

	dom := New(transport).Parse(html)
	dom.SetCookie("custom", "value123")

	jar := dom.GetCookieJar()
	if jar.cookies["custom"] != "value123" {
		t.Fatalf("expected cookie 'value123', got '%s'", jar.cookies["custom"])
	}

	// Submit a form and verify cookie is sent
	form, err := dom.Form("#test-form")
	if err != nil {
		t.Fatal(err)
	}

	_, err = form.Submit()
	if err != nil {
		t.Fatal(err)
	}

	// In a real scenario, we'd verify the cookie was sent in the request
	// For this test, we just verify it's stored
}

// Test BaseURL resolution
func TestBaseURLResolution(t *testing.T) {
	transport := NewMockTransport()
	transport.AddResponse("GET", "/about", 200, "About page")

	html := `<a id="about-link" href="/about">About</a>`

	dom := New(transport).Parse(html)

	// Set custom base URL
	_, err := dom.SetBaseURL("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	link, err := dom.Link("#about-link")
	if err != nil {
		t.Fatal(err)
	}

	linkURL, err := link.GetURL()
	if err != nil {
		t.Fatal(err)
	}

	expected := "http://example.com/about"
	actual := linkURL.String()
	if actual != expected {
		t.Fatalf("expected '%s', got '%s'", expected, actual)
	}
}

// Benchmark form operations
func BenchmarkFormFill(b *testing.B) {
	html := `
	<form id="test-form">
		<input type="text" name="field1">
		<input type="text" name="field2">
		<input type="text" name="field3">
	</form>`

	dom := New(NewMockTransport()).Parse(html)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		form, _ := dom.Form("#test-form")
		form.MustFill("field1", "value1").
			MustFill("field2", "value2").
			MustFill("field3", "value3")
	}
}

// SimpleMockTransport for example
type SimpleMockTransport struct{}

func (m SimpleMockTransport) Do(req Request) (Response, error) {
	if req.Method == http.MethodPost && req.URL.Path == "/login" {
		user := req.Form.Get("username")
		return Response{Status: 200, Body: "<p>Welcome, " + user + "</p>"}, nil
	}
	return Response{Status: 404, Body: "not found"}, nil
}

// Example from README
func ExampleDOM_Form() {
	html := `
	<form id="login-form" action="/login" method="post">
		<label>User</label><input type="text" name="username">
		<label>Pass</label><input type="password" name="password">
		<button type="submit">Login</button>
	</form>`

	dom := New(SimpleMockTransport{}).Parse(html)

	form, err := dom.Form("#login-form")
	if err != nil {
		fmt.Println(err)
		return
	}

	form.MustFill("username", "alice").MustFill("password", "secret")

	res, err := form.Submit()
	if err != nil {
		fmt.Println(err)
		return
	}

	if res.Status != 200 {
		fmt.Printf("expected status 200, got %d\n", res.Status)
		return
	}

	fmt.Println("Login successful")
	// Output: Login successful
}
