package main

const clientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"

const (
	gqlURL      = "https://gql.twitch.tv/gql"
	playlistURL = "https://usher.ttvnw.net/api/channel/hls/%s.m3u8"
)

const prefetchTag = "#EXT-X-TWITCH-PREFETCH:"
const mapTag = "#EXT-X-MAP:"
const infTag = "#EXTINF:"
const discontinuityTag = "#EXT-X-DISCONTINUITY"
const mediaSequenceTag = "#EXT-X-MEDIA-SEQUENCE:"
const endListTag = "#EXT-X-ENDLIST"

const maxSeenURLs = 50
