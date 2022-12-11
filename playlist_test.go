package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestGetPlaylist(t *testing.T) {
	testUsername := "testing"
	token := &accessToken{
		Value:     "token",
		Signature: "sig",
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, fmt.Sprintf(playlistURL, testUsername),
			req.URL.String()[:len(req.URL.String())-len(req.URL.RawQuery)-1])
		equals(t, "true", req.URL.Query().Get("allow_source"))
		equals(t, "true", req.URL.Query().Get("fast_bread"))
		equals(t, token.Signature, req.URL.Query().Get("sig"))
		equals(t, token.Value, req.URL.Query().Get("token"))

		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(bytes.NewBufferString(`#EXTM3U
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="chunked",NAME="1080p60 (source)",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1234567,RESOLUTION=1920x1080,CODECS="avc1.64002A,mp4a.40.2",VIDEO="chunked",FRAME-RATE=60.000
https://example.invalid/123.m3u8
#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID="720p60",NAME="720p60",AUTOSELECT=YES,DEFAULT=YES
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7654321,RESOLUTION=1280x720,CODECS="avc1.4D401F,mp4a.40.2",VIDEO="720p60",FRAME-RATE=60.000
https://example.invalid/456.m3u8`)),
			Header: make(http.Header),
		}
	})

	playlists, err := getPlaylists(client, testUsername, token)
	ok(t, err)

	equals(t, []playlistInfo{
		{
			Group:     "chunked",
			Name:      "1080p60 (source)",
			Bandwidth: 1234567,
			Width:     1920,
			Height:    1080,
			URL:       "https://example.invalid/123.m3u8",
		},
		{
			Group:     "720p60",
			Name:      "720p60",
			Bandwidth: 7654321,
			Width:     1280,
			Height:    720,
			URL:       "https://example.invalid/456.m3u8",
		},
	}, playlists)
}
