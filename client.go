package securehttp

import (
	"io"
	"net/http"
	"time"
)

type IClient interface {
	Do(url string, method string, timeout time.Duration, headers http.Header, body io.Reader) (*http.Response, error)
	DoForget(url string, method string, timeout time.Duration, headers http.Header, body io.Reader)
}

type clientConfig struct {
	netHttpClient *http.Client
}

func Setup(timeout time.Duration, customTransport *http.Transport) *clientConfig {
	// TODO: implement dnscache for better performance later
	// r := &dnscache.Resolver{}
	transporter := &http.Transport{
		// TODO: custom dial context combined with dnscache
		// DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
		// 	separator := strings.LastIndex(addr, ":")
		// 	ips, err := r.LookupHost(ctx, addr[:separator])
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	for _, ip := range ips {
		// 		conn, err = net.Dial(network, ip+addr[separator:])
		// 		if err == nil {
		// 			break
		// 		}
		// 	}
		// 	return
		// },
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          1024,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       10 * time.Second,
	}
	// TODO: dnscache background check and update
	// go func() {
	// 	clearUnused := true
	// 	t := time.NewTicker(5 * time.Minute)
	// 	defer t.Stop()
	// 	for range t.C {
	// 		r.Refresh(clearUnused)
	// 	}
	// }()

	if customTransport != nil {
		transporter = customTransport
	}

	return &clientConfig{
		netHttpClient: &http.Client{
			Transport: transporter,
			Timeout:   timeout,
		},
	}
}

func (c *clientConfig) buildReq(method, url string, body io.Reader, headers http.Header) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err == nil {
		for header, vals := range headers {
			for _, val := range vals {
				req.Header.Add(header, val)
			}
		}
	}

	return req, err
}

func (c *clientConfig) Do(url string, method string, timeout time.Duration, headers http.Header, body io.Reader) (*http.Response, error) {
	req, err := c.buildReq(method, url, body, headers)
	if err != nil {
		return nil, err
	}

	return c.netHttpClient.Do(req)
}

func (c *clientConfig) DoForget(url string, method string, timeout time.Duration, headers http.Header, body io.Reader) {
	go func() {
		resp, _ := c.Do(url, method, timeout, headers, body)
		if resp != nil {
			// Consume the entire body so we can reuse this connection
			defer resp.Body.Close()
			io.ReadAll(resp.Body)
		}
	}()
}
