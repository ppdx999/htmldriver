// Package chitransport provides a Transport implementation for Chi framework testing.
package chitransport

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	h "github.com/ppdx999/htmldriver"
)

// ChiTransport is a Transport implementation that routes requests through a Chi router
type ChiTransport struct {
	router  chi.Router
	baseURL string
}

// NewChiTransport creates a new ChiTransport with the given Chi router
func NewChiTransport(router chi.Router) *ChiTransport {
	return &ChiTransport{
		router:  router,
		baseURL: "http://localhost",
	}
}

// NewChiTransportWithBaseURL creates a new ChiTransport with a custom base URL
func NewChiTransportWithBaseURL(router chi.Router, baseURL string) *ChiTransport {
	return &ChiTransport{
		router:  router,
		baseURL: baseURL,
	}
}

// Do executes the request through the Chi router and returns the response
func (t *ChiTransport) Do(req h.Request) (h.Response, error) {
	// Build full URL
	fullURL := req.URL.String()
	if !req.URL.IsAbs() {
		baseURL, err := url.Parse(t.baseURL)
		if err != nil {
			return h.Response{}, err
		}
		fullURL = baseURL.ResolveReference(req.URL).String()
	}

	// Create HTTP request
	var body io.Reader
	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
		if len(req.Form) > 0 {
			body = strings.NewReader(req.Form.Encode())
		}
	}

	httpReq, err := http.NewRequest(req.Method, fullURL, body)
	if err != nil {
		return h.Response{}, err
	}

	// Copy headers
	for key, values := range req.Header {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	// Set Content-Type if not already set and we have form data
	if len(req.Form) > 0 && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Create response recorder
	recorder := httptest.NewRecorder()

	// Execute request through Chi router
	t.router.ServeHTTP(recorder, httpReq)

	// Read response body
	bodyBytes, err := io.ReadAll(recorder.Body)
	if err != nil {
		return h.Response{}, err
	}

	// Build response
	resp := h.Response{
		Header: recorder.Header(),
		URL:    httpReq.URL,
		Body:   string(bodyBytes),
		Status: recorder.Code,
	}

	return resp, nil
}

// SetBaseURL updates the base URL for the transport
func (t *ChiTransport) SetBaseURL(baseURL string) {
	t.baseURL = baseURL
}

// GetRouter returns the underlying Chi router
func (t *ChiTransport) GetRouter() chi.Router {
	return t.router
}
