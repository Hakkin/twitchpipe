package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	errNoLinks    = errors.New("no links found")
	errStreamOver = errors.New("stream over")
)

func getURLs(c *http.Client, playlist string) ([]string, error) {
	req, err := http.NewRequest("GET", playlist, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Origin", "https://player.twitch.tv")

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errStreamOver
		}

		return nil, fmt.Errorf("urls got http status %s", res.Status)
	}

	var urls []string

	var done bool
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		switch v := scanner.Text(); {
		case !strings.HasPrefix(v, "#"):
			urls = append(urls, v)
		case strings.HasPrefix(v, prefetchTag):
			urls = append(urls, v[len(prefetchTag):])
		case v == "#EXT-X-ENDLIST":
			done = true
		}
	}

	if len(urls) == 0 {
		return nil, errNoLinks
	}

	if done {
		err = errStreamOver
	}

	return urls, err
}
