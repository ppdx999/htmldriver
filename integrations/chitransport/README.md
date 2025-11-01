# chitransport

Chi framework integration for htmldriver.

## Installation

```bash
go get github.com/ppdx999/htmldriver/integrations/chitransport
```

## Usage

```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    h "github.com/ppdx999/htmldriver"
    "github.com/ppdx999/htmldriver/integrations/chitransport"
)

func main() {
    // Create Chi router
    r := chi.NewRouter()

    r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
        r.ParseForm()
        username := r.FormValue("username")
        w.Write([]byte("<p>Welcome, " + username + "</p>"))
    })

    // Create transport
    transport := chitransport.NewChiTransport(r)

    // Use with htmldriver
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

    // res.Status == 200
    // res.Body contains "Welcome, alice"
}
```

## Features

- Direct integration with Chi router
- No need for httptest.Server
- Automatic cookie handling
- Support for route parameters
- Custom base URL support

## API

### NewChiTransport

Creates a new ChiTransport with the given Chi router.

```go
func NewChiTransport(router chi.Router) *ChiTransport
```

### NewChiTransportWithBaseURL

Creates a new ChiTransport with a custom base URL.

```go
func NewChiTransportWithBaseURL(router chi.Router, baseURL string) *ChiTransport
```

### SetBaseURL

Updates the base URL for the transport.

```go
func (t *ChiTransport) SetBaseURL(baseURL string)
```

### GetRouter

Returns the underlying Chi router.

```go
func (t *ChiTransport) GetRouter() chi.Router
```
