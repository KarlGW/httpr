package httpr

import (
	"math/rand"
	"net/http"
	"time"
)

// Backoff is a function that provides delays between retries with backoff.
type Backoff func(minDelay, maxDelay time.Duration, jitter float64) time.Duration

// ShouldRetry is a function that evaluates if a retry should be done.
type ShouldRetry func(r *http.Response, err error) bool

// RetryPolicy contains rules for retries.
type RetryPolicy struct {
	// ShouldRetry is a function that evaluates if a retry should be done.
	ShouldRetry ShouldRetry
	// Backoff is a function that provides backoff between retries.
	Backoff Backoff
	// MaxRetries is the maximum amount of retries.
	MaxRetries int
	// MinDelay for the backoff.
	MinDelay time.Duration
	// MaxDelay for the backoff.
	MaxDelay time.Duration
	// Jitter for the backoff. A small number is recommended (0.1-0.2).
	Jitter float64
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

// ExponentialBackoff provides backoff with an increasing delay from min delay,
// to max delay with jitter.
func ExponentialBackoff() func(minDelay, maxDelay time.Duration, jitter float64) time.Duration {
	retry := 0
	return func(minDelay, maxDelay time.Duration, jitter float64) time.Duration {
		d := minDelay * (1 << retry)
		if d >= maxDelay {
			d = maxDelay
		}
		retry++

		if jitter > 0 {
			return d + calculateJitter(d, jitter)
		}
		return d
	}
}

// defaultRetryPolicy returns the default retry policy.
func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		ShouldRetry: StandardShouldRetry,
		Backoff:     ExponentialBackoff(),
		MaxRetries:  3,
		MinDelay:    500 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Jitter:      0.2,
	}
}

// calculateJitter calculates a random jitter value.
func calculateJitter(minDelay time.Duration, jitter float64) time.Duration {
	val := float64(minDelay) * jitter
	return time.Duration(float64(rand.Float64()*2*val - val))
}
