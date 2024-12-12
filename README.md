# httpr

[![Go Reference](https://pkg.go.dev/badge/github.com/KarlGW/httpr.svg)](https://pkg.go.dev/github.com/KarlGW/httpr)

> HTTP transport with retries

* [Getting started](#getting-started)
  * [Prerequisites](#prerequisites)
  * [Install](#install)
  * [Examples](#examples)
* [Usage](#usage)
  * [Retry policy](#retry-policy)
  * [Additional transports](#additional-transports)
  * [No retries](#no-retries)


This module provides a transport (`http.RoundTripper`) with retry functionality that can be used with the standard library HTTP client.

## Getting started

### Prerequisites

* Go 1.18

### Install

```
go get github.com/KarlGW/httpr
```

***Alternative***:

```sh
git clone git@github.com:KarlGW/httpr.git # git clone https://github.com/KarlGW/httpr.git
mkdir /path/to/project/httpr
cp httpr/*.go /path/to/project/httpr
```

### Examples

**Setting up the transport with the default retry policy:**

```go
package main

import (
    "net/http"
    "time"

    "github.com/KarlGW/httpr"
)

func main() {
    client := &http.Client{
        Transport: httpr.New(), // httpr.NewTransport also works for clarity.
    }

    req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
    if err != nil {
        // Handle error.
    }

    resp, err := client.Do(req)
    if err != nil {
        // Handle error.
    }
}
```

**Setting up the transport with a custom retry policy:**

```go
package main

import (
    "net/http"
    "time"

    "github.com/KarlGW/httpr"
)

func main() {
    client := &http.Client{
        Transport: httpr.New(httpr.WithRetryPolicy(
            httpr.RetryPolicy{
                ShouldRetry: httpr.StandardShouldRetry,
                Backoff:     httpr.ExponentialBackoff(),
                MaxRetries:  3,
                MinDelay:    500 * time.Milliseconds,
                MaxDelay:    5 * time.Seconds,
                Jitter:      0.2,
            }
        )),
    }

    req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
    if err != nil {
        // Handle error.
    }

    resp, err := client.Do(req)
    if err != nil {
        // Handle error.
    }
}
```

## Usage

### Retry policy

By default the `http.Transport` is configured with a default policy of:

```go
httpr.RetryPolicy{
    ShouldRetry: httpr.StandardShouldRetry,
    Backoff:     httpr.ExponentialBackoff(),
    MaxRetries:  3,
    MinDelay:    500 * time.Milliseconds,
    MaxDelay:    5 * time.Seconds,
    Jitter:      0.2,
}
```

Which means that it will retry on client errors and HTTP statuses `408`, `429`, `500`, `502`, `503` and `504`
with exponential backoff with 3 max retries. The backoff begins with 500 milliseconds and will double between each
attempt with a maximum final duration of 5 seconds. The jitter adds some randomness to the backoff.

### Additional transports

If another transport (`http.RoundTripper`) should be used as the underlying transport (transport chaining), it can
be provided with the option `WithTransport`.
