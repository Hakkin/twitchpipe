package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type accessToken struct {
	Token string `json:"token"`
	Sig   string `json:"sig"`
}

func getAcessToken(c *http.Client, username string) (*accessToken, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(accessURL, username), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Client-ID", clientID)

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusNotFound {
			return nil, fmt.Errorf("access got http status %s", res.Status)
		}
		return nil, errors.New("stream does not exist")
	}

	var token accessToken
	err = json.NewDecoder(res.Body).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
