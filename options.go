package httpr

import "net/http"

// WithRetryPolicy sets the retry policy for the transport.
func WithRetryPolicy(retryPolicy RetryPolicy) Option {
	return func(t *Transport) {
		if !retryPolicy.IsZero() {
			t.rp = retryPolicy
		}
	}
}

// WithTransport sets the underlying transport. Use when
// other custom transports are needed.
func WithTransport(transport http.RoundTripper) Option {
	return func(t *Transport) {
		if t != nil {
			t.tr = transport
		}
	}
}
