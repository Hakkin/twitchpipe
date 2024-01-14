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
				fmt.Sprintf("#EXTM3U\n#EXT-X-MAP:URI=\"%s\"\n#EXTINF:2.00,live\n%s\n#EXT-X-TWITCH-PREFETCH:%s\n", initURL, normalURL, prefetchURL),
			)),
			Header: make(http.Header),
		}
	})

	urls, err := getURLs(client, "https://example.invalid/123.m3u8")
	ok(t, err)

	equals(t, 2, len(urls))
	equals(t, Segment{
		Name:          "",
		URI:           normalURL,
		MapURI:        initURL,
		Duration:      0,
		Seq:           0,
		Discontinuity: false,
		Prefetch:      false,
	}, urls[0])
	equals(t, Segment{
		Name:          "",
		URI:           prefetchURL,
		MapURI:        initURL,
		Duration:      0,
		Seq:           1,
		Discontinuity: false,
		Prefetch:      true,
	}, urls[1])
}
