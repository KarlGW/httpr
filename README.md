# httpr

[![Go Reference](https://pkg.go.dev/badge/github.com/KarlGW/httpr.svg)](https://pkg.go.dev/github.com/KarlGW/httpr)

> HTTP transport with retries

* [Getting started](#getting-started)
  * [Prerequisites](#prerequisites)
  * [Install](#install)
  * [Example](#example)


This module provides a transport (`http.RoundTripper`) with retry functionality that can be used with the standard library HTTP client.

## Getting started

### Prerequisites

* Go 1.18

### Install

```
go get github.com/KarlGW/httpr
```

### Example

Setting up the transport with the default retry policy:

```go
package main

import (
    "net/http"

    "github.com/KarlGW/httpr"
)

func main() {
    client := &http.Client{
        Transport: httpr.New(), // httpr.NewTransport also works for clarity.
    }

    req, err := http.NewRequest(http.MethodGet, "http://example.com")
    if err != nil {
        // Handle error.
    }

    resp, err := client.Do(req)
    if err != nil {
        // Handle error.
    }
}
```
