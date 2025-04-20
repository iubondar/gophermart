package testhelpers

import "errors"

// ErrorReader implements io.Reader and always returns an error.
// Useful for testing error handling in functions that read from io.Reader.
type ErrorReader struct{}

func (er *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}
