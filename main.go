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

	"golang.org/x/term"
	"rsc.io/getopt"
)

var stdErr = log.New(os.Stderr, "", 0)

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

	username = strings.ToLower(username)

	if term.IsTerminal(int(os.Stdout.Fd())) && !forceOutput && !externalCommand && !groupList {
		stdErr.Println("[WARNING] You have not piped the output anywhere.")
		stdErr.Println("          Outputting binary data to a terminal can be dangerous.")
		stdErr.Println("          To bypass this safety feature, use the '--force-output' option.")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	variables := map[string]any{
		"platform":   accessTokenPlatform,
		"playerType": accessTokenPlayerType,
	}
	if accessTokenPlayerBackend.string != nil {
		variables["playerBackend"] = *accessTokenPlayerBackend.string
	}

	token, err := getAcessToken(client, username, accessTokenOAuth.string, accessTokenDeviceID.string, variables)
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

	var currentSeq int
	var needInit bool = true
	urls, urlsErr := getURLs(client, playlistURL)
	if !archiveMode && len(urls) > 1 {
		currentSeq = urls[len(urls)-1].Seq
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
			if url.Seq < currentSeq {
				continue
			}

			if url.Discontinuity {
				needInit = true
			}

			if url.MapURI != "" && needInit {
				tsURLs <- url.MapURI
			}
			needInit = false

			tsURLs <- url.URI

			currentSeq = url.Seq + 1
		}

		time.Sleep(time.Second * 1)

		urls, urlsErr = getURLs(client, playlistURL)
	}
}
