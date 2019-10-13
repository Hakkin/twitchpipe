package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetURLs(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, "https://example.invalid/123.m3u8", req.URL.String())
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(prefetchTag + "https://example.invalid/123.ts")),
			Header:     make(http.Header),
		}
	})

	prefetch, err := getURLs(client, "https://example.invalid/123.m3u8")
	ok(t, err)

	equals(t, 1, len(prefetch))
	equals(t, "https://example.invalid/123.ts", prefetch[0])
}
