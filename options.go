package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"rsc.io/getopt"
)

type optionalString struct {
	*string
}

func (o *optionalString) Set(s string) error {
	if o.string == nil {
		o.string = new(string)
	}

	*o.string = s

	return nil
}

func (o *optionalString) String() string {
	if o.string == nil {
		return "unset"
	}

	return *o.string
}

const version = "1.0"

var (
	forceOutput        bool
	forceOutputDefault = false

	usernameURL        bool
	usernameURLDefault = false

	archiveMode        bool
	archiveModeDefault = false

	hideConsole        bool
	hideConsoleDefault = false

	groupSelect        string
	groupSelectDefault = "best"

	groupList        bool
	groupListDefault = false

	showVersion        bool
	showVersionDefault = false

	accessTokenPlatform        string
	accessTokenPlatformDefault = "web"

	accessTokenPlayerType        string
	accessTokenPlayerTypeDefault = "site"

	accessTokenPlayerBackend optionalString
	accessTokenOAuth         optionalString
	accessTokenDeviceID      optionalString
)

func init() {
	flag.BoolVar(&forceOutput, "f", forceOutputDefault, "Force output to standard output even if TTY is detected")
	flag.BoolVar(&usernameURL, "u", usernameURLDefault, "Treat USERNAME as a URL")
	flag.BoolVar(&archiveMode, "a", archiveModeDefault, "Start downloading from the oldest segment rather than the newest")
	flag.StringVar(&groupSelect, "g", groupSelectDefault, "Select specified playlist group\n\t\"best\" will select the best available group")
	flag.BoolVar(&groupList, "G", groupListDefault, "List available playlist groups and exit")
	flag.BoolVar(&showVersion, "v", showVersionDefault, "Show version information and exit")
	getopt.Aliases(
		"f", "force-output",
		"u", "url",
		"a", "archive",
		"g", "group",
		"G", "list-groups",
		"v", "version",
	)

	flag.StringVar(&accessTokenPlatform, "access-token-platform", accessTokenPlatformDefault, "The platform to send when acquiring an access token")
	flag.StringVar(&accessTokenPlayerType, "access-token-player-type", accessTokenPlayerTypeDefault, "The player type to send when acquiring an access token")
	flag.Var(&accessTokenPlayerBackend, "access-token-player-backend", "The player backend to send when acquiring an access token (optional)")
	flag.Var(&accessTokenOAuth, "access-token-oauth", "OAuth token to send when acquiring an access token (optional)")
	flag.Var(&accessTokenDeviceID, "access-token-device-id", "Device ID to send when acquiring an access token (optional)")

	// Check if the version flag is set
	flag.Parse()
	if showVersion {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}
}

func printUsage() {
	stdErr.Println("Usage: twitchpipe [OPTIONS...] <USERNAME> [COMMAND...]")
	stdErr.Println()
	stdErr.Println("If COMMAND is specified, it will be executed and stream data will be \nwritten to its standard input.")
	stdErr.Println("Otherwise, stream data will be written to standard output.")
	stdErr.Println()
	stdErr.Println("Options:")
	getopt.PrintDefaults()
}

func findBest(playlists []playlistInfo) playlistInfo {
	var best playlistInfo
	var highBitrate int
	for _, p := range playlists {
		if p.Group == "chunked" {
			return p
		}

		if p.Bandwidth > highBitrate {
			highBitrate = p.Bandwidth
			best = p
		}
	}

	return best
}

func printGroups(playlists []playlistInfo) {
	columns := []*struct {
		title   string
		length  int
		content []string
		fn      func(p playlistInfo) string
	}{
		{"Group", 0, nil, func(p playlistInfo) string { return p.Group }},
		{"Name", 0, nil, func(p playlistInfo) string { return p.Name }},
		{"Resolution", 0, nil, func(p playlistInfo) string { return fmt.Sprintf("%dx%d", p.Width, p.Height) }},
		{"Codec", 0, nil, func(p playlistInfo) string {
			var cs string
			for _, c := range strings.Split(p.Codec, ",") {
				cs += strings.SplitN(c, ".", 2)[0] + "+"
			}
			if cs != "" {
				cs = cs[:len(cs)-1]
			}
			return cs
		}},
		{"Bitrate", 0, nil, func(p playlistInfo) string { return fmt.Sprintf("%dk", p.Bandwidth/1024) }},
	}

	for _, c := range columns {
		c.length = len(c.title)
	}

	for _, p := range playlists {
		for _, c := range columns {
			content := c.fn(p)
			c.content = append(c.content, content)
			if len(content) > c.length {
				c.length = len(content)
			}
		}
	}

	for _, c := range columns {
		fmt.Fprint(os.Stderr, c.title)
		if c.length-len(c.title) > 0 {
			fmt.Fprint(os.Stderr, strings.Repeat(" ", c.length-len(c.title)))
		}
		fmt.Fprint(os.Stderr, " ")
	}

	fmt.Fprintln(os.Stderr)

	best := findBest(playlists)

	for i := range playlists {
		for _, c := range columns {
			content := c.content[i]
			fmt.Fprint(os.Stderr, content)
			if c.length-len(content) > 0 {
				fmt.Fprint(os.Stderr, strings.Repeat(" ", c.length-len(content)))
			}
			fmt.Fprint(os.Stderr, " ")
		}

		if playlists[i].Group == best.Group {
			fmt.Fprint(os.Stderr, "(best)")
		}

		fmt.Fprintln(os.Stderr)
	}
}
