package httpr_test

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/KarlGW/httpr"
)

func ExampleNew() {
	// Setup an HTTP client with the httpr Transport with the default retry policy.
	client := &http.Client{
		Transport: httpr.New(),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		// Handle error.
	}

	resp, err := client.Do(req)
	if err != nil {
		// Handle err.
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		// Handle err.
	}
	fmt.Println(string(b))
}

func ExampleNew_withOptions() {
	// Setup an HTTP client with the httpr Transport with a custom retry policy.
	client := &http.Client{
		Transport: httpr.New(httpr.WithRetryPolicy(httpr.RetryPolicy{
			ShouldRetry: httpr.StandardShouldRetry,
			Backoff:     httpr.ExponentialBackoff(),
			MaxRetries:  3,
			MinDelay:    500 * time.Millisecond,
			MaxDelay:    5 * time.Second,
			Jitter:      0.2,
		})),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		// Handle error.
	}

	resp, err := client.Do(req)
	if err != nil {
		// Handle err.
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		// Handle err.
	}
	fmt.Println(string(b))
}

func ExampleNewTransport() {
	// Setup an HTTP client with the httpr Transport with the default retry policy.
	client := &http.Client{
		Transport: httpr.NewTransport(),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		// Handle error.
	}

	resp, err := client.Do(req)
	if err != nil {
		// Handle err.
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		// Handle err.
	}
	fmt.Println(string(b))
}

func ExampleNewTransport_withOptions() {
	// Setup an HTTP client with the httpr Transport with a custom retry policy.
	client := &http.Client{
		Transport: httpr.NewTransport(httpr.WithRetryPolicy(httpr.RetryPolicy{
			ShouldRetry: httpr.StandardShouldRetry,
			Backoff:     httpr.ExponentialBackoff(),
			MaxRetries:  3,
			MinDelay:    500 * time.Millisecond,
			MaxDelay:    5 * time.Second,
			Jitter:      0.2,
		})),
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		// Handle error.
	}

	resp, err := client.Do(req)
	if err != nil {
		// Handle err.
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		// Handle err.
	}
	fmt.Println(string(b))
}

func ExampleTransport_create() {
	// Setup an HTTP client with the httpr Transport. When using a struct
	// literal the default retry policy is used.
	client := &http.Client{
		Transport: &httpr.Transport{},
	}

	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		// Handle error.
	}

	resp, err := client.Do(req)
	if err != nil {
		// Handle err.
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		// Handle err.
	}
	fmt.Println(string(b))
}
