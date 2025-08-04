package securehttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

type IClient interface {
	Get() *http.Client
	MultipartWithNethttp(opts MultipartRequestOptions) error
	DoWithNethttp(opts RequestOptions) error
}

type clientConfig struct {
	clientNetHttp *http.Client
}

func (c *clientConfig) Get() *http.Client {
	return c.clientNetHttp
}

// newTransport configures a custom HTTP transport for connection pooling and timeouts.
func newTransport() *http.Transport {
	return &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func NewClient() *clientConfig {
	return &clientConfig{
		clientNetHttp: &http.Client{
			Transport: newTransport(),
			Timeout:   30 * time.Second,
		},
	}
}

func (c *clientConfig) MultipartWithNethttp(opts MultipartRequestOptions) error {
	// determine timeout
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	// create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// prepare the request body reader
	var bodyReader io.Reader
	if opts.Multipart != nil && opts.Body.Bytes() != nil {
		bodyReader = bytes.NewReader(opts.Body.Bytes())
	}

	// build the HTTP request
	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// set multipart content type if needed
	if opts.Multipart != nil {
		req.Header.Set("Content-Type", opts.Multipart.FormDataContentType())
	}

	// set any additional headers
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// perform the request
	resp, err := c.clientNetHttp.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// check expected status code
	if resp.StatusCode != opts.ExpectCode {
		// read up to 1KB of body to include in the error
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		return fmt.Errorf("unexpected status code: got %d, want %d; body: %q",
			resp.StatusCode, opts.ExpectCode, string(buf[:n]))
	}

	// decode JSON output if requested
	if opts.Out != nil {
		if err := json.NewDecoder(resp.Body).Decode(opts.Out); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
		}
	}

	return nil
}

func (c *clientConfig) DoWithNethttp(opts RequestOptions) error {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var reqBody []byte
	if opts.Body != nil {
		jsonBody, err := json.Marshal(opts.Body)
		if err != nil {
			return fmt.Errorf("marshal body failed: %w", err)
		}

		reqBody = jsonBody
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := c.clientNetHttp.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != opts.ExpectCode {
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		return fmt.Errorf("unexpected status code: got %d, want %d; body: %q",
			resp.StatusCode, opts.ExpectCode, string(buf[:n]))
	}

	if opts.Out != nil {
		if err := json.NewDecoder(resp.Body).Decode(opts.Out); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
		}
	}

	return nil
}
