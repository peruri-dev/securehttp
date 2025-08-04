package securehttp

import (
	"bytes"
	"mime/multipart"
	"time"
)

type RequestOptions struct {
	Method      string
	URL         string
	QueryParams map[string]string
	Body        any
	Headers     map[string]string
	ExpectCode  int
	Out         any
	Timeout     time.Duration
}

type MultipartRequestOptions struct {
	Method      string
	URL         string
	QueryParams map[string]string
	Body        bytes.Buffer
	Headers     map[string]string
	ExpectCode  int
	Out         any
	Timeout     time.Duration
	Multipart   *multipart.Writer
}

type Build struct {
	Data   any   `json:"data"`
	Errors []any `json:"errors"`
	Meta   any   `json:"meta"`
}

type ErrResponse struct {
	ID     string       `json:"id"`
	Status int          `json:"status"`
	Code   string       `json:"code"`
	Title  string       `json:"title"`
	Detail string       `json:"detail"`
	Source *ErrorSource `json:"source,omitempty"`
	Meta   any          `json:"meta,omitempty"`
}

type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
	Header    string `json:"header,omitempty"`
}
