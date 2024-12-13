package httpr

import (
	"io"
	"net/http"
	"time"
)

// Transport is a transport that provides retry functionality.
type Transport struct {
	tr http.RoundTripper
	rp RetryPolicy
}

// Option is a function that configures the Transport.
type Option func(t *Transport)

// New creates and configures a new transport. If no
// retry policy is provided a default one will be set.
//
// The default retry policy has retries with exponential backoff with
// the foloowing values:
//
// MaxRetries: 3, MinDelay: 500ms, MaxDelay: 5s and Jitter: 0.2
//
// It retries on client errors and HTTP statuses:
// 408, 429, 500, 502, 503 and 504.
func New(options ...Option) *Transport {
	tr := &Transport{}
	for _, option := range options {
		option(tr)
	}
	if tr.tr == nil {
		tr.tr = http.DefaultTransport
	}
	if tr.rp.IsZero() {
		tr.rp = defaultRetryPolicy()
	}
	return tr
}

// NewTansport creates and configures a new transport. If no
// retry policy is provided a default one will be set.
//
// The default retry policy has retries with exponential backoff with
// the foloowing values:
//
// MaxRetries: 3, MinDelay: 500ms, MaxDelay: 5s and Jitter: 0.2
//
// It retries on client errors and HTTP statuses:
// 408, 429, 500, 502, 503 and 504.
var NewTransport = New

// RoundTrip satisfies the http.RoundTripper interface and performs an
// http request with the configured retry policy.
func (tr *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if tr.tr == nil {
		tr.tr = http.DefaultTransport
	}
	if tr.rp.IsZero() {
		tr.rp = defaultRetryPolicy()
	}
	if tr.rp.ShouldRetry == nil {
		tr.rp.ShouldRetry = func(r *http.Response, err error) bool {
			return false
		}
	}
	if tr.rp.Backoff == nil {
		tr.rp.Backoff = func(minDelay, maxDelay time.Duration, jitter float64) time.Duration {
			return 0
		}
	}
	backoff := tr.rp.Backoff

	retries := 0
	for {
		resp, err := tr.tr.RoundTrip(r)
		if !tr.rp.ShouldRetry(resp, err) || retries >= tr.rp.MaxRetries {
			return resp, err
		}

		delay := backoff(tr.rp.MinDelay, tr.rp.MaxDelay, tr.rp.Jitter)
		select {
		case <-time.After(delay):
			retries++
			if err := drainResponse(resp); err != nil {
				return nil, err
			}
			if _, err := resetRequest(r); err != nil {
				return nil, err
			}
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}
	}
}

// Set configures the transport with the provided options.
func (tr *Transport) Set(options ...Option) {
	for _, option := range options {
		option(tr)
	}
}

// resetRequest resets the request to be used before a retry.
func resetRequest(r *http.Request) (*http.Request, error) {
	req := r.Clone(r.Context())
	if r.Body != nil && r.GetBody != nil {
		body, err := r.GetBody()
		if err != nil {
			return nil, err
		}
		req.Body = body
	}
	r = req
	return r, nil
}

// drainResponse drains the response body before a retry.
func drainResponse(r *http.Response) error {
	if r == nil {
		return nil
	}
	defer r.Body.Close()
	if _, err := io.Copy(io.Discard, r.Body); err != nil {
		return err
	}
	return nil
}
