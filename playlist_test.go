package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetPlaylist(t *testing.T) {
	testUsername := "testing"
	token := &accessToken{
		Token: "token",
		Sig:   "sig",
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, fmt.Sprintf(playlistURL, testUsername),
			req.URL.String()[:len(req.URL.String())-len(req.URL.RawQuery)-1])
		equals(t, "true", req.URL.Query().Get("allow_source"))
		equals(t, "true", req.URL.Query().Get("fast_bread"))
		equals(t, token.Sig, req.URL.Query().Get("sig"))
		equals(t, token.Token, req.URL.Query().Get("token"))

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString("https://example.invalid/123.m3u8")),
			Header:     make(http.Header),
		}
	})

	playlist, err := getPlaylist(client, testUsername, token)
	ok(t, err)

	equals(t, playlist, "https://example.invalid/123.m3u8")
}
