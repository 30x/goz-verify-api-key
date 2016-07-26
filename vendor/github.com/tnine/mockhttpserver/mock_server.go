package mockhttpserver

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
)

//MockServer a server that will perform requests
type MockServer struct {
	responses []*Response

	//stoppable hook for server once started
	server *graceful.Server

	//the address we're listening on
	address string
}

/**
* Relationship of structure is NewRequest().Headers
**/

//Request The root type of request that should be created
type Request struct {
	//the verb of the request
	verb string
	//the Url of the request
	url string
	//the body of the request that's expected.  Can be nil
	body []byte
	//the list of required headers
	header []header

	//mockServer The server this request needs added to
	mockServer *MockServer
}

//Header The value that must be present in the header
type header struct {
	name  string
	value string
}

//The Response type for constructing responses
type Response struct {
	//the status to return
	status int

	//The body to return in the stream, can be nil for no boxy
	body []byte

	//All headers that will be returned
	header []header

	//request The request for this response
	request *Request
}

//NewRequest create a new request
func (mockServer *MockServer) NewRequest(verb string, url string, body []byte) *Request {
	return &Request{
		verb:       verb,
		url:        url,
		body:       body,
		header:     []header{},
		mockServer: mockServer,
	}
}

//NewGet convenience method for new GET
func (mockServer *MockServer) NewGet(url string) *Request {
	return mockServer.NewRequest("GET", url, nil)
}

//NewPost convenience method for new NewPost
func (mockServer *MockServer) NewPost(url string, contentType string, body []byte) *Request {
	return mockServer.NewRequest("POST", url, body).AddHeader("Content-Type", contentType)
}

//AddHeader return ourselves so we can continue to add headers
func (request *Request) AddHeader(name string, value string) *Request {
	request.header = append(request.header, header{name: name, value: value})

	return request
}

//ToResponse Add the response to the request
func (request *Request) ToResponse(status int, body []byte) *Response {
	return &Response{
		status:  status,
		body:    body,
		request: request,
	}
}

//AddHeader return ourselves so we can continue to add headers
func (response *Response) AddHeader(name string, value string) *Response {
	response.header = append(response.header, header{name: name, value: value})

	return response
}

//Add this request/response to our list
func (response *Response) Add() {
	response.request.mockServer.responses = append(response.request.mockServer.responses, response)
}

//Listen set up the routes and listen.  Address can be of the format "127.0.0.1:8080"
func (mockServer *MockServer) Listen(address string) error {

	//wire the routes
	r := mux.NewRouter()

	for _, response := range mockServer.responses {

		request := response.request

		route := r.Methods(request.verb).Path(request.url)

		//if we have headers add them
		for _, header := range request.header {
			route = route.HeadersRegexp(header.name, header.value)
		}

		expectedBody := []byte{}

		// TODO, we need to group by URL and add switch matching for content body
		//Set up our handler function to respond.
		route.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			//if we have a body make sure it's equal.

			log.Println("Incoming request for", req.URL.Path)

			if len(expectedBody) > 0 {

				passedBody, err := ioutil.ReadAll(req.Body)

				if err != nil {
					http.Error(w, "Unable to read body", http.StatusInternalServerError)
					return
				}

				defer req.Body.Close()

				equals := reflect.DeepEqual(expectedBody, passedBody)

				if !equals {

					expected := string(expectedBody)
					actual := string(passedBody)

					resposneBody := fmt.Sprintf("Expected body of \n===============\n %s \n===============\n\n\n.  However received body of \n===============\n %s \n===============\n", expected, actual)

					http.Error(w, resposneBody, http.StatusInternalServerError)
					return
				}

			}

			//we've set the headers
			for _, header := range response.header {
				w.Header().Add(header.name, header.value)
			}

			//write the status
			w.WriteHeader(response.status)

			if len(response.body) > 0 {
				_, err := w.Write(response.body)

				if err != nil {
					errorString := fmt.Sprintf("Unable to write response body on method %s and url %s.  Error is %s", request.verb, request.url, err)

					fmt.Println(errorString)

					http.Error(w, errorString, http.StatusInternalServerError)
					return
				}
			}

		})

	}

	//now log every request to stout
	loggedRouter := handlers.CombinedLoggingHandler(os.Stdout, r)

	log.Printf("Starting server at address %s", address)

	mockServer.server = &graceful.Server{
		Timeout: 1 * time.Millisecond,
		Server:  &http.Server{Addr: address, Handler: loggedRouter},
		Logger:  graceful.DefaultLogger(),
	}

	//start listening
	if err := mockServer.server.ListenAndServe(); err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			return err
		}
	}

	return nil

}

//Shutdown the test server
func (mockServer *MockServer) Shutdown() error {

	mockServer.server.Stop(1 * time.Millisecond)

	return testSocket(mockServer.address)
}

//StartAsync starts the server in the background.  Ensure you defer server.Shutdown() after invoking this
func (mockServer *MockServer) StartAsync(address string) error {

	go func() {
		mockServer.Listen(address)
	}()

	err := testSocket(address)

	if err != nil {
		return err
	}

	return nil
}

func testSocket(address string) error {
	started := make(chan bool)

	go func() {

		for i := 0; i < 20; i++ {

			conn, err := net.Dial("tcp", address)

			//done waiting, continue
			if err == nil {
				conn.Close()
				started <- true
				break
			}

			time.Sleep(100 * time.Millisecond)
		}

		close(started)
	}()

	//break on which happens first, we're started of we time out
	select {
	case <-started:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("Timed out after 5 seconds")
	}

}

func writeServerError(error string) {

}