package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var (
	errNoLinks    = errors.New("no links found")
	errStreamOver = errors.New("stream over")
)

var initRegex = regexp.MustCompile(`URI="([^"]*)"`)

type segmentURL struct {
	URL    string
	IsInit bool
}

func getURLs(c *http.Client, playlist string) ([]segmentURL, error) {
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

	var urls []segmentURL

	var done bool
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		switch v := scanner.Text(); {
		case !strings.HasPrefix(v, "#"):
			urls = append(urls, segmentURL{v, false})
		case strings.HasPrefix(v, prefetchTag):
			urls = append(urls, segmentURL{v[len(prefetchTag):], false})
		case strings.HasPrefix(v, initTag):
			matches := initRegex.FindStringSubmatch(v[len(initTag):])
			if len(matches) != 2 {
				break
			}
			urls = append(urls, segmentURL{matches[1], true})
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
