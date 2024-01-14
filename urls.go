package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errNoLinks    = errors.New("no links found")
	errStreamOver = errors.New("stream over")
)

var initRegex = regexp.MustCompile(`URI="([^"]*)"`)

type Segment struct {
	Name          string
	URI           string
	MapURI        string
	Duration      float64
	Seq           int
	Discontinuity bool
	Prefetch      bool
}

type DateRange struct {
	ID        string
	Class     string
	StartDate time.Time
	EndDate   time.Time
	Duration  time.Duration
	EndOnNext bool
	Extra     []string
}

func getURLs(c *http.Client, playlist string) ([]Segment, error) {
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

	var urls []Segment

	var done bool
	var segment Segment
	var mapURI string
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		switch v := scanner.Text(); {
		case strings.HasPrefix(v, mapTag):
			matches := initRegex.FindStringSubmatch(v[len(mapTag):])
			if len(matches) != 2 {
				break
			}
			mapURI = matches[1]
		case strings.HasPrefix(v, mediaSequenceTag):
			sequenceText := v[len(mediaSequenceTag):]
			seq, err := strconv.Atoi(sequenceText)
			if err != nil {
				break
			}
			segment.Seq = seq
		case v == discontinuityTag:
			segment.Discontinuity = true
		case v == endListTag:
			done = true
		case strings.HasPrefix(v, prefetchTag):
			v = v[len(prefetchTag):]
			segment.Prefetch = true
			fallthrough
		case !strings.HasPrefix(v, "#"):
			if len(urls) > 0 {
				segment.Seq = urls[len(urls)-1].Seq + 1
			}
			segment.URI = v
			segment.MapURI = mapURI
			urls = append(urls, segment)
			segment = Segment{}
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
