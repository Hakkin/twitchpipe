package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetAccessToken(t *testing.T) {
	testUsername := "testing"
	testOAuth := "secret"
	testDeviceID := "abc123"
	accessToken := accessToken{
		Value:     "token",
		Signature: "sig",
	}

	gqlVariables := map[string]any{
		"testing": "123",
		"hello":   "world",
	}

	var gqlat gqlAccessToken
	gqlat.StreamPlaybackAccessToken.accessToken = accessToken

	var gqlrd bytes.Buffer
	ok(t, json.NewEncoder(&gqlrd).Encode(&gqlat))

	gqlr := gqlResponse{
		Data: gqlrd.Bytes(),
	}

	client := NewTestClient(func(req *http.Request) *http.Response {
		equals(t, gqlURL, req.URL.String())
		equals(t, http.MethodPost, req.Method)
		equals(t, clientID, req.Header.Get("Client-ID"))
		equals(t, "OAuth "+testOAuth, req.Header.Get("Authorization"))
		equals(t, testDeviceID, req.Header.Get("Device-ID"))

		var gqlq gqlQuery
		ok(t, json.NewDecoder(req.Body).Decode(&gqlq))

		equals(t, testUsername, gqlq.Variables["channelName"])
		for k := range gqlVariables {
			equals(t, gqlVariables[k], gqlq.Variables[k])
		}

		var jsonBuf bytes.Buffer
		ok(t, json.NewEncoder(&jsonBuf).Encode(&gqlr))
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&jsonBuf),
			Header:     make(http.Header),
		}
	})

	testToken, err := getAcessToken(client, testUsername, &testOAuth, &testDeviceID, gqlVariables)
	ok(t, err)
	equals(t, &accessToken, testToken)
}
