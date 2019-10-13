package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func getPlaylist(c *http.Client, username string, token *accessToken) (string, error) {
	pURL, err := url.Parse(fmt.Sprintf(playlistURL, username))
	if err != nil {
		return "", err
	}

	query := pURL.Query()

	query.Set("allow_source", "true")
	query.Set("fast_bread", "true")
	query.Set("sig", token.Sig)
	query.Set("token", token.Token)

	pURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", pURL.String(), nil)
	if err != nil {
		return "", err
	}

	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusNotFound {
			return "", fmt.Errorf("playlist got http status %s", res.Status)
		}
		return "", errors.New("stream is offline")
	}

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	var url string
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "http") {
			url = scanner.Text()
			break
		}
	}

	if url == "" {
		return "", errors.New("no http links in playlist")
	}

	return url, nil
}
