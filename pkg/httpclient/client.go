package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// DefaultRoundTripper is used if no RoundTripper is set in Config.
var DefaultRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 10 * time.Second,
}

// Config defines configuration parameters for a new client.
type Config struct {
	// The address to connect to.
	Address string

	// RoundTripper is used by the Client to drive HTTP requests. If not
	// provided, DefaultRoundTripper will be used.
	RoundTripper http.RoundTripper
}

func (cfg *Config) roundTripper() http.RoundTripper {
	if cfg.RoundTripper == nil {
		return DefaultRoundTripper
	}
	return cfg.RoundTripper
}

// Client is the interface for an API client.
type Client interface {
	URL(ep string, args map[string]string) *url.URL
	Do(context.Context, *http.Request, interface{}) (*http.Response, []byte, error)
}

// NewClient returns a new Client.
//
// It is safe to use the returned Client from multiple goroutines.
func NewClient(cfg Config) (Client, error) {
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimRight(u.Path, "/")

	return &httpClient{
		endpoint: u,
		client:   http.Client{Transport: cfg.roundTripper()},
	}, nil
}

type httpClient struct {
	endpoint *url.URL
	client   http.Client
}

func (c *httpClient) URL(ep string, args map[string]string) *url.URL {
	p := path.Join(c.endpoint.Path, ep)

	for arg, val := range args {
		arg = ":" + arg
		p = strings.Replace(p, arg, val, -1)
	}

	u := *c.endpoint
	u.Path = p

	return &u
}

func (c *httpClient) Do(ctx context.Context, req *http.Request, result interface{}) (*http.Response, []byte, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := c.client.Do(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, nil, err
	}

	var body []byte
	done := make(chan struct{})
	go func() {
		body, err = ioutil.ReadAll(resp.Body)
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		err = resp.Body.Close()
		if err == nil {
			err = fmt.Errorf("context timeout / %w", ctx.Err())
		}
	case <-done:
		var msg string
		if resp.StatusCode >= 400 {
			msg = fmt.Sprintf("(%d) %s", resp.StatusCode, string(body))
		}

		switch {
		case resp.StatusCode >= 500:
			err = &ErrInternal{msg}
		case resp.StatusCode == 404:
			err = &ErrNotFound{msg}
		case resp.StatusCode >= 400:
			err = &ErrClient{msg}
		}
	}

	if err != nil {
		return resp, body, err
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return resp, body, fmt.Errorf("error unmarshal response body / %w", err)
		}
	}

	return resp, body, nil
}
