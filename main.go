package main

import (
	"flag"
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
)

func init() {
	flag.BoolVar(&forceOutput, "f", false, "Force output to standard output even if TTY is detected")
	flag.BoolVar(&usernameURL, "u", false, "Treat USERNAME as a URL")
	flag.BoolVar(&archiveMode, "a", false, "Start downloading from the oldest segment rather than the newest")
	getopt.Aliases(
		"f", "force-output",
		"u", "url",
		"a", "archive",
	)
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

	if terminal.IsTerminal(int(os.Stdout.Fd())) && !forceOutput && !externalCommand {
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

	playlistURL, err := getPlaylist(client, username, token)
	if err != nil {
		stdErr.Printf("could not extract playlist: %v\n", err)
		os.Exit(1)
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
