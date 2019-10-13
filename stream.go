package main

import (
	"fmt"
	"io"
	"net/http"
)

func streamTs(c *http.Client, ts <-chan string, out io.Writer, done chan<- error) {
	for url := range ts {
		for {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				done <- fmt.Errorf("couldn't create ts request: %w", err)
				return
			}

			res, err := c.Do(req)
			if err != nil {
				done <- fmt.Errorf("couldn't get ts: %w", err)
				return
			}

			if res.StatusCode < 200 || res.StatusCode >= 300 {
				stdErr.Printf("got non-2xx http status %s, skipping segment\n", res.Status)
				break
			}

			_, err = io.Copy(out, res.Body)
			if err != nil {
				if err == io.ErrUnexpectedEOF {
					stdErr.Printf("got unexpected EOF while copying ts to output, skipping segment\n")
					break
				}

				done <- fmt.Errorf("couldn't copy ts to output: %w", err)
				return
			}

			res.Body.Close()
			break
		}
	}
	done <- nil
}
