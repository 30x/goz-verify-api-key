package pipeline

import "net/http"

// ResponseHandlerFunc is the part of the Pipe that handles responses
type ResponseHandlerFunc func(w http.ResponseWriter, r *http.Request, res *http.Response)
