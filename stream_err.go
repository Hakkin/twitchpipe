package main

import (
	"fmt"
	"io"
)

type writeError struct {
	Err error
}

func (err *writeError) Error() string {
	return fmt.Sprintf("write: %v", err.Err)
}
func (err *writeError) Unwrap() error {
	return err.Err
}

type writerError struct {
	w io.Writer
}

func (w *writerError) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		return n, &writeError{err}
	}
	return n, nil
}

type readError struct {
	Err error
}

func (err *readError) Error() string {
	return fmt.Sprintf("read: %v", err.Err)
}
func (err *readError) Unwrap() error {
	return err.Err
}

type readerError struct {
	r io.Reader
}

func (r *readerError) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if err != nil {
		return n, &readError{err}
	}
	return n, nil
}

type retryError struct {
	Err error
}

func (err *retryError) Error() string {
	return fmt.Sprintf("retrying segment: %v", err.Err)
}
func (err *retryError) Unwrap() error {
	return err.Err
}

type skipError struct {
	Err error
}

func (err *skipError) Error() string {
	return fmt.Sprintf("skipping segment: %v", err.Err)
}
func (err *skipError) Unwrap() error {
	return err.Err
}

type fatalError struct {
	Err error
}

func (err *fatalError) Error() string {
	return fmt.Sprintf("fatal error: %v", err.Err)
}
func (err *fatalError) Unwrap() error {
	return err.Err
}
