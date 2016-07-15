package pipeline

import (
	"net/http"
	"runtime/debug"
)

// ErrorHandlerFunc is a function called to handle an error in a Pipe
type ErrorHandlerFunc func(writer http.ResponseWriter, err error) error

// DefaultErrorHanderFunc for now just sends the error to the client
func DefaultErrorHanderFunc(writer http.ResponseWriter, err error) error {
	writer.WriteHeader(500)
	_, err = writer.Write([]byte(err.Error() + "\n"))
	if err != nil {
		_, err = writer.Write(debug.Stack())
	}
	return err
}
