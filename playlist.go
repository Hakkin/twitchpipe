package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type playlistInfo struct {
	Name      string
	Group     string
	Bandwidth int
	Width     int
	Height    int
	URL       string
}

var (
	groupRegex      = regexp.MustCompile(`GROUP-ID="([^"]+)"`)
	nameRegex       = regexp.MustCompile(`NAME="([^"]+)"`)
	bandwidthRegex  = regexp.MustCompile(`BANDWIDTH=([0-9]+)`)
	resolutionRegex = regexp.MustCompile(`RESOLUTION=([0-9]+x[0-9]+)`)
)

func getPlaylists(c *http.Client, username string, token *accessToken) ([]playlistInfo, error) {
	pURL, err := url.Parse(fmt.Sprintf(playlistURL, username))
	if err != nil {
		return nil, err
	}

	query := pURL.Query()

	query.Set("allow_source", "true")
	query.Set("allow_audio_only", "true")
	query.Set("fast_bread", "true")
	query.Set("sig", token.Signature)
	query.Set("token", token.Value)

	pURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", pURL.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode != http.StatusNotFound {
			return nil, fmt.Errorf("playlist got http status %s", res.Status)
		}
		return nil, errors.New("stream is offline")
	}

	scanner := bufio.NewScanner(res.Body)
	scanner.Split(bufio.ScanLines)

	var playlists []playlistInfo

	var info playlistInfo
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "http") {
			info.URL = scanner.Text()
			playlists = append(playlists, info)
			info = playlistInfo{}
			continue
		}

		if !strings.HasPrefix(scanner.Text(), "#EXT") {
			continue
		}

		split := strings.SplitN(scanner.Text(), ":", 2)
		if len(split) != 2 {
			continue
		}

		tag, attribute := split[0], split[1]
		switch tag {
		case "#EXT-X-MEDIA":
			groupMatch := groupRegex.FindStringSubmatch(attribute)
			if groupMatch != nil {
				info.Group = groupMatch[1]
			}

			nameMatch := nameRegex.FindStringSubmatch(attribute)
			if nameMatch != nil {
				info.Name = nameMatch[1]
			}
		case "#EXT-X-STREAM-INF":
			bandwidthMatch := bandwidthRegex.FindStringSubmatch(attribute)
			if bandwidthMatch != nil {
				bandwidthInt, err := strconv.Atoi(bandwidthMatch[1])
				if err == nil {
					info.Bandwidth = bandwidthInt
				}
			}

			resolutionMatch := resolutionRegex.FindStringSubmatch(attribute)
			if resolutionMatch != nil {
				resolution := strings.SplitN(resolutionMatch[1], "x", 2)
				if len(resolution) == 2 {
					width, widthErr := strconv.Atoi(resolution[0])
					height, heightErr := strconv.Atoi(resolution[1])
					if widthErr == nil && heightErr == nil {
						info.Width = width
						info.Height = height
					}
				}
			}
		}
	}

	return playlists, nil
}
