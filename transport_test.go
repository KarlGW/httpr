package httpr

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNew(t *testing.T) {
	var tests = []struct {
		name  string
		input []Option
		want  *Transport
	}{
		{
			name: "default transport",
			want: &Transport{
				tr: http.DefaultTransport,
				rp: defaultRetryPolicy(),
			},
		},
		{
			name: "with retry policy",
			input: []Option{
				WithRetryPolicy(RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  5,
					MinDelay:    1 * time.Second,
					MaxDelay:    10 * time.Second,
					Jitter:      0.1,
				}),
			},
			want: &Transport{
				tr: http.DefaultTransport,
				rp: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  5,
					MinDelay:    1 * time.Second,
					MaxDelay:    10 * time.Second,
					Jitter:      0.1,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := New(test.input...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Transport{}, http.Transport{}), cmpopts.IgnoreUnexported(http.Transport{}), cmpopts.IgnoreFields(RetryPolicy{}, "ShouldRetry", "Backoff"), cmpopts.IgnoreFields(http.Transport{}, "Proxy", "DialContext")); diff != "" {
				t.Errorf("New() = unexpected result (-want +got)\n%s\n", diff)
			}
		})
	}
}

func TestTransport_RoundTrip(t *testing.T) {
	type input struct {
		req         func() *http.Request
		retryPolicy RetryPolicy
		retries     int
		err         error
	}

	type want struct {
		statusCode int
		body       []byte
	}

	var tests = []struct {
		name    string
		input   input
		want    want
		wantErr error
	}{
		{
			name: "successful GET",
			input: input{
				req: func() *http.Request {
					req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				body:       wantBodyGet,
			},
		},
		{
			name: "successful GET with retries",
			input: input{
				req: func() *http.Request {
					req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
				retries: 3,
				err:     errors.New("error"),
			},
			want: want{
				statusCode: http.StatusOK,
				body:       wantBodyGet,
			},
		},
		{
			name: "failure GET with retries",
			input: input{
				req: func() *http.Request {
					req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
				retries: 4,
				err:     errors.New("error"),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       []byte(``),
			},
		},
		{
			name: "failure GET context exceeded",
			input: input{
				req: func() *http.Request {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
					defer cancel()
					req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
				retries: 4,
				err:     errors.New("error"),
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				body:       []byte(``),
			},
			wantErr: errors.New("error"),
		},
		{
			name: "successful POST",
			input: input{
				req: func() *http.Request {
					req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(wantBodyPost))
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
			},
			want: want{
				statusCode: http.StatusOK,
				body:       wantBodyPost,
			},
		},
		{
			name: "successful POST with retries",
			input: input{
				req: func() *http.Request {
					req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(wantBodyPost))
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
				retries: 3,
				err:     errors.New("error"),
			},
			want: want{
				statusCode: http.StatusOK,
				body:       wantBodyPost,
			},
		},
		{
			name: "successful POST with retries - request literal",
			input: input{
				req: func() *http.Request {
					u, _ := url.Parse("http://example.com")
					req := &http.Request{
						Method: http.MethodPost,
						URL:    u,
						Body:   io.NopCloser(bytes.NewReader(wantBodyPost)),
					}
					return req
				},
				retryPolicy: RetryPolicy{
					ShouldRetry: StandardShouldRetry,
					Backoff:     ExponentialBackoff(),
					MaxRetries:  3,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
				},
				retries: 3,
				err:     errors.New("error"),
			},
			want: want{
				statusCode: http.StatusOK,
				body:       wantBodyPost,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := setupServer(false, test.input.retries, test.input.err)
			defer ts.Close()
			client := setupClient(ts.Listener.Addr().String(), test.input.retryPolicy, test.input.err)

			got, gotErr := client.Do(test.input.req())
			if test.wantErr == nil && got == nil {
				t.Fatalf("error in test")
			}

			if test.wantErr != nil {
				if gotErr == nil {
					t.Errorf("RoundTrip() = unexpected result, should return an error\n")
				}
				return
			}

			if test.want.statusCode != got.StatusCode {
				t.Errorf("RoundTrip() = unexpected result, want: %d, got: %d\n", test.want.statusCode, got.StatusCode)
			}

			gotBody, _ := io.ReadAll(got.Body)
			defer got.Body.Close()

			if diff := cmp.Diff(test.want.body, gotBody); diff != "" {
				t.Errorf("RoundTrip() = unexpected result (-want +got)\n%s\n", diff)
			}
		})
	}
}

func setupServer(tls bool, retries int, err error) *httptest.Server {
	type newTestServerFunc func(handler http.Handler) *httptest.Server
	var newFn newTestServerFunc
	if tls {
		newFn = httptest.NewTLSServer
	} else {
		newFn = httptest.NewServer
	}

	count := 0
	return newFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil && count < retries {
			count++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			b, _ := io.ReadAll(r.Body)
			defer r.Body.Close()
			w.Write(b)
			return
		}
		w.Write(wantBodyGet)
	}))
}

func setupClient(target string, rp RetryPolicy, err error) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Dial: func(network, addr string) (net.Conn, error) {
			if err != nil && errors.Is(err, errTestClient) {
				return nil, err
			}
			return net.Dial("tcp", target)
		},
	}
	return &http.Client{
		Transport: New(WithRetryPolicy(rp), WithTransport(tr)),
	}
}

var wantBodyGet = []byte(`{"message":"hello"}`)
var wantBodyPost = []byte(`{"message":"post"}`)
var errTestClient = errors.New("client error")
