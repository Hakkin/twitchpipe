package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	testUsername := "testing"
	token := &accessToken{
		Token: "token",
		Sig:   "sig",
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, fmt.Sprintf(accessURL, testUsername), req.URL.String())
		equals(t, clientID, req.Header.Get("Client-ID"))
		var jsonBuf bytes.Buffer
		ok(t, json.NewEncoder(&jsonBuf).Encode(token))
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(&jsonBuf),
			Header:     make(http.Header),
		}
	})

	testToken, err := getAcessToken(client, testUsername)
	ok(t, err)
	equals(t, token, testToken)
}
