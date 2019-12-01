package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
	"rsc.io/getopt"
)

var stdErr = log.New(os.Stderr, "", 0)

var (
	forceOutput bool
	usernameURL bool
	archiveMode bool
	hideConsole bool
	groupSelect string
	groupList   bool
)

func init() {
	flag.BoolVar(&forceOutput, "f", false, "Force output to standard output even if TTY is detected")
	flag.BoolVar(&usernameURL, "u", false, "Treat USERNAME as a URL")
	flag.BoolVar(&archiveMode, "a", false, "Start downloading from the oldest segment rather than the newest")
	flag.StringVar(&groupSelect, "g", "best", "Select specified playlist group\n\t\"best\" will select the best available group")
	flag.BoolVar(&groupList, "G", false, "List available playlist groups and exit")
	getopt.Aliases(
		"f", "force-output",
		"u", "url",
		"a", "archive",
		"g", "group",
		"G", "list-groups",
	)
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

func printUsage() {
	stdErr.Println("Usage: twitchpipe [OPTIONS...] <USERNAME> [COMMAND...]")
	stdErr.Println()
	stdErr.Println("If COMMAND is specified, it will be executed and stream data will be \nwritten to its standard input.")
	stdErr.Println("Otherwise, stream data will be written to standard output.")
	stdErr.Println()
	stdErr.Println("Options:")
	getopt.PrintDefaults()
}

func main() {
	getopt.Parse()
	if len(flag.Args()) < 1 {
		printUsage()
		os.Exit(1)
	}

	if hideConsole {
		hideWindow()
	}

	externalCommand, externalArgs := len(flag.Args()) > 1, len(flag.Args()) > 2

	username := flag.Arg(0)
	if usernameURL {
		u, err := url.Parse(username)
		if err != nil {
			stdErr.Fatalf("could not parse username as URL: %v\n", err)
		}

		path := u.Path
		path = strings.TrimPrefix(path, "/")

		username = strings.Split(path, "/")[0]
	}

	if username == "" {
		printUsage()
		os.Exit(1)
	}

	if terminal.IsTerminal(int(os.Stdout.Fd())) && !forceOutput && !externalCommand && !groupList {
		stdErr.Println("[WARNING] You have not piped the output anywhere.")
		stdErr.Println("          Outputting binary data to a terminal can be dangerous.")
		stdErr.Println("          To bypass this safety feature, use the '--force-output' option.")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	token, err := getAcessToken(client, username)
	if err != nil {
		stdErr.Printf("could not acquire access token: %v\n", err)
		os.Exit(1)
	}

	playlists, err := getPlaylists(client, username, token)
	if err != nil {
		stdErr.Printf("could not extract playlist: %v\n", err)
		os.Exit(1)
	}

	if groupList {
		printGroups(playlists)
		os.Exit(0)
	}

	var playlistURL string
	switch groupSelect {
	case "best":
		best := findBest(playlists)
		playlistURL = best.URL
	default:
		for _, p := range playlists {
			if p.Group == groupSelect {
				playlistURL = p.URL
				break
			}
		}
	}

	if playlistURL == "" {
		stdErr.Printf("could not find desired playlist quality")
		os.Exit(2)
	}

	var output io.Writer = os.Stdout
	if len(flag.Args()) > 1 {
		var args []string
		if externalArgs {
			args = flag.Args()[2:]
		}
		cmd := exec.Command(flag.Arg(1), args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if output, err = cmd.StdinPipe(); err != nil {
			stdErr.Fatalf("could not acquire external command input: %v\n", err)
		}

		if err = cmd.Start(); err != nil {
			stdErr.Fatalf("could not start external command: %v\n", err)
		}
		defer cmd.Wait()
	}

	tsURLs := make(chan string, 2)
	done := make(chan error, 1)
	go streamTs(client, tsURLs, output, done)

	var seenURLs []string
	seenURLsIndex := make(map[string]bool)

	urls, urlsErr := getURLs(client, playlistURL)
	if !archiveMode {
		if len(urls) > 1 {
			for _, url := range urls[0 : len(urls)-1] {
				seenURLsIndex[url] = true
				seenURLs = append(seenURLs, url)
			}
		}
	}

	for {
		select {
		case err := <-done:
			close(tsURLs)
			stdErr.Printf("error while streaming: %v\n", err)
			os.Exit(2)
		default:
		}

		if urlsErr != nil {
			if urlsErr == errStreamOver {
				close(tsURLs)
				err := <-done
				if err != nil {
					stdErr.Printf("stream over with error: %v\n", err)
					os.Exit(2)
				}
				stdErr.Println("stream over")
				os.Exit(0)
			}

			stdErr.Printf("could not get prefetch URLs: %v\n", urlsErr)
		}

		for _, url := range urls {
			if seenURLsIndex[url] {
				continue
			}
			seenURLsIndex[url] = true
			seenURLs = append(seenURLs, url)

			tsURLs <- url
		}

		if len(seenURLs) > maxSeenURLs {
			var removed []string
			delta := len(seenURLs) - maxSeenURLs
			removed, seenURLs = seenURLs[0:delta-1], seenURLs[delta:]
			for _, url := range removed {
				delete(seenURLsIndex, url)
			}
		}

		time.Sleep(time.Second * 2)

		urls, urlsErr = getURLs(client, playlistURL)
	}
}
