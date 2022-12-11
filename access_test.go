package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	testUsername := "testing"
	token := &accessToken{
		Token: "token",
		Sig:   "sig",
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		aURL, err := url.Parse(fmt.Sprintf(accessURL, testUsername))
		ok(t, err)

		query := aURL.Query()
		query.Set("player_type", "embed")
		aURL.RawQuery = query.Encode()

		equals(t, aURL.String(), req.URL.String())
		equals(t, clientID, req.Header.Get("Client-ID"))
		var jsonBuf bytes.Buffer
		ok(t, json.NewEncoder(&jsonBuf).Encode(token))
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&jsonBuf),
			Header:     make(http.Header),
		}
	})

	testToken, err := getAcessToken(client, testUsername)
	ok(t, err)
	equals(t, token, testToken)
}
