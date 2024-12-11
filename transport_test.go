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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

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
					ShouldRetry: defaultShouldRetry,
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
					ShouldRetry: defaultShouldRetry,
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
					ShouldRetry: defaultShouldRetry,
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
					ShouldRetry: defaultShouldRetry,
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
					ShouldRetry: defaultShouldRetry,
					Backoff:     defaultBackoff,
					MinDelay:    1 * time.Millisecond,
					MaxDelay:    5 * time.Millisecond,
					MaxRetries:  3,
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
