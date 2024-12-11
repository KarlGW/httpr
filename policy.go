package httpr

import (
	"math"
	"net/http"
	"time"
)

// Backoff is a function that provides delays between retries with backoff.
type Backoff func(delay, maxDelay time.Duration, retry int) time.Duration

// Retry is a function that evaluates if a retry should be done.
type Retry func(r *http.Response, err error) bool

// RetryPolicy contains rules for retries.
type RetryPolicy struct {
	Retry      Retry
	Backoff    Backoff
	MinDelay   time.Duration
	MaxDelay   time.Duration
	MaxRetries int
}

// IsZero returns true if the RetryPolicy is the zero value.
func (r RetryPolicy) IsZero() bool {
	return r.Retry == nil && r.Backoff == nil && r.MinDelay == 0 && r.MaxDelay == 0 && r.MaxRetries == 0
}

// standardRetry is the standard implementation of retry.
func standardRetry(r *http.Response, err error) bool {
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

// defaultRetry is the default implementation of retry.
var defaultRetry = standardRetry

// exponentialBackoff provides backoff with an increasing delay from min delay,
// to max delay.
func exponentialBackoff(delay, maxDelay time.Duration, retry int) time.Duration {
	d := delay * time.Duration(math.Pow(2, float64(retry)))
	if d >= maxDelay {
		d = maxDelay
	}
	return d
}

// defaultBackoff is the default implementation of backoff.
var defaultBackoff = exponentialBackoff

// defaultRetryPolicy returns the default retry policy.
func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		Backoff:    defaultBackoff,
		Retry:      defaultRetry,
		MinDelay:   500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		MaxRetries: 3,
	}
}
