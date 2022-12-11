package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
)

type accessToken struct {
	Value     string `json:"value"`
	Signature string `json:"signature"`
}

type gqlQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type gqlAccessToken struct {
	StreamPlaybackAccessToken struct {
		accessToken
	} `json:"streamPlaybackAccessToken"`
}

//go:embed access_token.gql
var accessTokenQuery string

func getAcessToken(c *http.Client, channelName string, oAuthToken *string, deviceID *string, variables map[string]any) (*accessToken, error) {
	variables["channelName"] = channelName
	q := &gqlQuery{
		Query:     accessTokenQuery,
		Variables: variables,
	}

	qs, err := json.Marshal(q)
	if err != nil {
		return nil, fmt.Errorf("error marshalling GQL query string: %w", err)
	}

	req, err := http.NewRequest("POST", gqlURL, bytes.NewReader(qs))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", clientID)

	if oAuthToken != nil {
		req.Header.Set("Authorization", "OAuth "+*oAuthToken)
	}

	if deviceID != nil {
		req.Header.Set("Device-ID", *deviceID)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got non-200 http status code %s while fetching access token", res.Status)
	}

	var gqlRes gqlResponse
	if err = json.NewDecoder(res.Body).Decode(&gqlRes); err != nil {
		return nil, fmt.Errorf("error decoding GQL response: %w", err)
	}

	if len(gqlRes.Errors) != 0 {
		var gqlErr string
		for i := range gqlRes.Errors {
			gqlErr += gqlRes.Errors[i].Message
		}

		return nil, fmt.Errorf("GQL returned error(s) while fetching access token: %s", gqlErr)
	}

	var accessToken gqlAccessToken
	if err = json.Unmarshal(gqlRes.Data, &accessToken); err != nil {
		return nil, fmt.Errorf("error decoding access token: %w", err)
	}

	return &accessToken.StreamPlaybackAccessToken.accessToken, nil
}
