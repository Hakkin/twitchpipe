package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

var urlRegexp = regexp.MustCompile(`https?://[^\s]+\.ts$`)

var (
	errNoLinks    = errors.New("no links found")
	errStreamOver = errors.New("stream over")
)

func getURLs(c *http.Client, playlist string) ([]string, error) {
	req, err := http.NewRequest("GET", playlist, nil)
	if err != nil {
		return nil, err
	}

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
		if scanner.Text() == "#EXT-X-ENDLIST" {
			done = true
		}
		if url := urlRegexp.FindString(scanner.Text()); url != "" {
			urls = append(urls, url)
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
