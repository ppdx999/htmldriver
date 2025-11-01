package htmldriver

import (
	"net/http"
	"net/url"
)

// Request represents an HTTP request to be sent via Transport
type Request struct {
	Method string
	URL    *url.URL
	Header http.Header
	Form   url.Values
	Files  []FormFile
}

// FormFile represents a file to be uploaded in a multipart form
type FormFile struct {
	FieldName string
	FileName  string
	Content   []byte
}

// Response represents an HTTP response received from Transport
type Response struct {
	Status int
	Body   string
	Header http.Header
	URL    *url.URL
}

// Transport is an interface that abstracts HTTP communication
type Transport interface {
	Do(req Request) (Response, error)
}
