package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestStreamTs(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, "https://example.invalid/123.ts", req.URL.String())
		return &http.Response{
			StatusCode: 200,
			Body: &ReadCloserMock{
				ReadCloser: io.NopCloser(bytes.NewBufferString("CONTENTS")),
				CloserFunc: func() error {
					return nil
				},
			},
			Header: make(http.Header),
		}
	})

	ts := make(chan string)
	done := make(chan error)
	var out bytes.Buffer
	go streamTs(client, ts, &out, done)
	ts <- "https://example.invalid/123.ts"
	close(ts)
	err := <-done
	equals(t, nil, err)
	equals(t, "CONTENTS", out.String())
}
