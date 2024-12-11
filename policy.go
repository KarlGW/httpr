package httpr

import (
	"math"
	"net/http"
	"time"
)

// Backoff is a function that provides delays between retries with backoff.
type Backoff func(delay, maxDelay time.Duration, retry int) time.Duration

// ShouldRetry is a function that evaluates if a retry should be done.
type ShouldRetry func(r *http.Response, err error) bool

// RetryPolicy contains rules for retries.
type RetryPolicy struct {
	ShouldRetry ShouldRetry
	Backoff     Backoff
	MinDelay    time.Duration
	MaxDelay    time.Duration
	MaxRetries  int
}

// IsZero returns true if the RetryPolicy is the zero value.
func (r RetryPolicy) IsZero() bool {
	return r.ShouldRetry == nil && r.Backoff == nil && r.MinDelay == 0 && r.MaxDelay == 0 && r.MaxRetries == 0
}

// StandardShouldRetry is a standard implementation of ShouldRetry.
// It will retry on client errors and on response errors
// with status codes: 408, 429, 500, 502, 503 and 504.
func StandardShouldRetry(r *http.Response, err error) bool {
	if err != nil {
		return true
	}
	switch r.StatusCode {
	case 0:
		return true
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return true
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}

// defaultShouldRetry is the default implementation of ShouldRetry.
var defaultShouldRetry = StandardShouldRetry

// ExponentialBackoff provides backoff with an increasing delay from min delay,
// to max delay.
func ExponentialBackoff(minDelay, maxDelay time.Duration, retry int) time.Duration {
	d := minDelay * time.Duration(math.Pow(2, float64(retry)))
	if d >= maxDelay {
		d = maxDelay
	}
	return d
}

// defaultBackoff is the default implementation of Backoff.
var defaultBackoff = ExponentialBackoff

// defaultRetryPolicy returns the default retry policy.
func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		Backoff:     defaultBackoff,
		ShouldRetry: defaultShouldRetry,
		MinDelay:    500 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		MaxRetries:  3,
	}
}
