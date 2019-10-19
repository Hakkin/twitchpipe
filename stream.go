package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func streamTs(c *http.Client, ts <-chan string, out io.Writer, done chan<- error) {
	for url := range ts {
		for {
			err := func() error {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					return &retryError{fmt.Errorf("couldn't create ts request: %w", err)}
				}

				res, err := c.Do(req)
				if err != nil {
					return &retryError{fmt.Errorf("couldn't get ts: %w", err)}
				}
				defer res.Body.Close()

				if res.StatusCode < 200 || res.StatusCode >= 300 {
					return &skipError{fmt.Errorf("got non-2xx http status %s", res.Status)}
				}

				_, err = io.Copy(&writerError{out}, &readerError{res.Body})
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					if wErr, ok := err.(*writeError); ok {
						return &fatalError{fmt.Errorf("error while writing ts to output: %w", wErr.Unwrap())}
					}
					if errors.Is(err, io.ErrUnexpectedEOF) {
						return &skipError{fmt.Errorf("got unexpected EOF while copying ts to output")}
					}

					return &skipError{fmt.Errorf("couldn't copy ts to output: %w", err)}
				}

				return nil
			}()
			if err != nil {
				if _, ok := err.(*fatalError); ok {
					done <- err
					for range ts {
					}
					return
				}

				stdErr.Printf("%v\n", err)
				if _, ok := err.(*skipError); ok {
					break
				}
				if _, ok := err.(*retryError); ok {
					continue
				}
			}

			break
		}
	}
	done <- nil
}
