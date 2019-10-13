package main

const clientID = "jzkbprff40iqj646a697cyrvl0zt2m6"

const (
	accessURL   = "https://api.twitch.tv/api/channels/%s/access_token"
	playlistURL = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
)

const prefetchTag = "#EXT-X-TWITCH-PREFETCH:"

const maxSeenURLs = 50
