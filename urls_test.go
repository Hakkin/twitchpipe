package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestGetURLs(t *testing.T) {
	prefetchURL := "https://example.invalid/123.ts"
	initURL := "https://example.invalid/init.mp4"
	normalURL := "https://example.invalid/456.ts"

	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, "https://example.invalid/123.m3u8", req.URL.String())
		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(bytes.NewBufferString(
				fmt.Sprintf("#EXTM3U\n#EXT-X-MAP:URI=\"%s\"\n%s\n#EXT-X-TWITCH-PREFETCH:%s\n", initURL, normalURL, prefetchURL),
			)),
			Header: make(http.Header),
		}
	})

	urls, err := getURLs(client, "https://example.invalid/123.m3u8")
	ok(t, err)

	equals(t, 3, len(urls))
	equals(t, segmentURL{initURL, true}, urls[0])
	equals(t, segmentURL{normalURL, false}, urls[1])
	equals(t, segmentURL{prefetchURL, false}, urls[2])
}
