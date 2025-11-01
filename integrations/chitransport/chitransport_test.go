package chitransport

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	h "github.com/ppdx999/htmldriver"
)

func TestChiTransport_BasicRequest(t *testing.T) {
	// Create Chi router
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Home Page</h1>"))
	})

	// Create transport
	transport := NewChiTransport(r)

	// Create DOM and parse
	dom := h.New(transport).Parse("<a id=\"home\" href=\"/\">Home</a>")

	link, err := dom.Link("#home")
	if err != nil {
		t.Fatal(err)
	}

	res, err := link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	if !strings.Contains(res.Body, "Home Page") {
		t.Fatalf("expected 'Home Page' in response, got: %s", res.Body)
	}
}

func TestChiTransport_FormSubmission(t *testing.T) {
	// Create Chi router
	r := chi.NewRouter()

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "alice" && password == "secret" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<p>Welcome, " + username + "</p>"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("<p>Invalid credentials</p>"))
		}
	})

	// Create transport
	transport := NewChiTransport(r)

	// Create login form HTML
	html := `
	<form id="login-form" action="/login" method="post">
		<input type="text" name="username">
		<input type="password" name="password">
		<button type="submit">Login</button>
	</form>`

	dom := h.New(transport).Parse(html)

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

func TestChiTransport_WithCookies(t *testing.T) {
	// Create Chi router
	r := chi.NewRouter()

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "session",
			Value: "abc123",
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<a id="profile" href="/profile">Profile</a>`))
	})

	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "abc123" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("<p>Unauthorized</p>"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>Profile Page</h1>"))
	})

	// Create transport
	transport := NewChiTransport(r)

	// Login
	loginHTML := `<form id="login" action="/login" method="post"></form>`

	jar := h.NewCookieJar()
	dom := h.NewWithCookieJar(transport, jar).Parse(loginHTML)

	form, err := dom.Form("#login")
	if err != nil {
		t.Fatal(err)
	}

	res, err := form.Submit()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	// Click profile link with shared cookie jar
	profileDOM := h.NewWithCookieJar(transport, jar).Parse(res.Body)

	link, err := profileDOM.Link("#profile")
	if err != nil {
		t.Fatal(err)
	}

	res, err = link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	if !strings.Contains(res.Body, "Profile Page") {
		t.Fatalf("expected 'Profile Page' in response, got: %s", res.Body)
	}
}

func TestChiTransport_RouteParameters(t *testing.T) {
	// Create Chi router
	r := chi.NewRouter()

	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<h1>User " + id + "</h1>"))
	})

	// Create transport
	transport := NewChiTransport(r)

	// Create link to user page
	html := `<a id="user-link" href="/users/123">User 123</a>`

	dom := h.New(transport).Parse(html)

	link, err := dom.Link("#user-link")
	if err != nil {
		t.Fatal(err)
	}

	res, err := link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	if !strings.Contains(res.Body, "User 123") {
		t.Fatalf("expected 'User 123' in response, got: %s", res.Body)
	}
}

func TestChiTransport_CustomBaseURL(t *testing.T) {
	// Create Chi router
	r := chi.NewRouter()

	r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create transport with custom base URL
	transport := NewChiTransportWithBaseURL(r, "https://api.example.com")

	html := `<a id="api-link" href="/api/data">API</a>`

	dom := h.New(transport).Parse(html)

	link, err := dom.Link("#api-link")
	if err != nil {
		t.Fatal(err)
	}

	res, err := link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 200 {
		t.Fatalf("expected status 200, got %d", res.Status)
	}

	if !strings.Contains(res.Body, `"status":"ok"`) {
		t.Fatalf("expected JSON response, got: %s", res.Body)
	}
}

func TestChiTransport_404(t *testing.T) {
	// Create Chi router with no routes
	r := chi.NewRouter()

	// Create transport
	transport := NewChiTransport(r)

	html := `<a id="missing" href="/nonexistent">Missing</a>`

	dom := h.New(transport).Parse(html)

	link, err := dom.Link("#missing")
	if err != nil {
		t.Fatal(err)
	}

	res, err := link.Click()
	if err != nil {
		t.Fatal(err)
	}

	if res.Status != 404 {
		t.Fatalf("expected status 404, got %d", res.Status)
	}
}

// Example from README
func ExampleNewChiTransport() {
	r := chi.NewRouter()

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.FormValue("username")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<p>Welcome, " + username + "</p>"))
	})

	transport := NewChiTransport(r)

	html := `
	<form id="login-form" action="/login" method="post">
		<input type="text" name="username">
		<input type="password" name="password">
		<button type="submit">Login</button>
	</form>`

	dom := h.New(transport).Parse(html)

	form, _ := dom.Form("#login-form")
	form.MustFill("username", "alice").MustFill("password", "secret")

	res, _ := form.Submit()

	if res.Status == 200 {
		// Success
	}
}
