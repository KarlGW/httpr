package httpr

import "net/http"

// WithRetryPolicy sets the retry policy for the transport.
func WithRetryPolicy(retryPolicy RetryPolicy) Option {
	return func(tr *Transport) {
		if !retryPolicy.IsZero() {
			tr.rp = retryPolicy
		}
	}
}

// WithTransport sets the underlying transport. Use when
// other custom transports are needed.
func WithTransport(transport http.RoundTripper) Option {
	return func(tr *Transport) {
		if tr != nil {
			tr.tr = transport
		}
	}
}

// WithNoRetries configures the transport to not perform any retries.
func WithNoRetries() Option {
	return func(tr *Transport) {
		tr.rp.ShouldRetry = func(r *http.Response, err error) bool {
			return false
		}
	}
}
