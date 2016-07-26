package verifyAPIKey

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/30x/gozerian/pipeline"
	"github.com/30x/gozerian/test_util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tnine/mockhttpserver"
)

const organization = "radical-new"
const environment = "test"
const headerKey = "X-Apigee-API-Key"

var _ = Describe("Test Mock Server", func() {

	It("No api key ", func() {
		//use the pipe fitting to create the pipeline
		pipeline := createPipeline()

		//get our request handler
		requestHandler := pipeline.RequestHandlerFunc()

		//create a mock reqeust.
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if err != nil {
			log.Fatal(err)
		}

		w := httptest.NewRecorder()

		requestHandler(w, req)

		Expect(w.Code).Should(Equal(http.StatusUnauthorized), "No api key was presented")

	})

	It("Not Found api key ", func() {

		const headerKey = "X-Apigee-API-Key"
		const headerValue = "TestValue"

		//use the pipe fitting to create the pipeline
		pipeline := createPipeline()

		//get our request handler
		requestHandler := pipeline.RequestHandlerFunc()

		//Mock up our apid server
		server := &mockhttpserver.MockServer{}

		//mock up the apid server and start it

		form := url.Values{}
		form.Add("action", "verify")
		form.Add("organization", organization)
		form.Add("environment", environment)
		form.Add("key", headerValue)
		form.Add("uriPath", "/foo")

		formAsString := form.Encode()

		formBytes := []byte(formAsString)

		failureResponse := `
		{
                "responseType": "APIKeyContext",
                "responseCode": 404,
                "result": {
                  "errorCode": "FOO_AUTHENTICATION_FAILED",
                  "reason": "APIKey expired"
                }
              }`

		//post to the endpoint, then return a 404
		server.NewPost("/verifiers/apikey", "application/x-www-form-urlencoded", formBytes).ToResponse(http.StatusOK, []byte(failureResponse)).Add()

		//start the server
		err := server.StartAsync("127.0.0.1:8181")

		defer server.Shutdown()

		Expect(err).Should(BeNil(), "Server should start")

		//create a mock reqeust.
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if err != nil {
			log.Fatal(err)
		}

		//add the api key
		req.Header.Add(headerKey, "foo")

		w := httptest.NewRecorder()

		requestHandler(w, req)

		//when apid returns a 404, the plugin should return a 401
		Expect(w.Code).Should(Equal(http.StatusUnauthorized), "No api key was presented")

	})

	It("Valid api key ", func() {

		const headerKey = "X-Apigee-API-Key"
		const headerValue = "TestValue"

		//use the pipe fitting to create the pipeline
		pipeline := createPipeline()

		//get our request handler
		requestHandler := pipeline.RequestHandlerFunc()

		//Mock up our apid server
		server := &mockhttpserver.MockServer{}

		//mock up the apid server and start it

		form := url.Values{}
		form.Add("action", "verify")
		form.Add("organization", organization)
		form.Add("environment", environment)
		form.Add("key", headerValue)
		form.Add("uriPath", "/foo")

		formAsString := form.Encode()

		formBytes := []byte(formAsString)

		template := `
		{
                "responseType": "APIKeyContext",
                "responseCode": 200,
                "result": {
                  "key": "%s",
                  "expiresAt": 1234567890,
                  "issuedAt": 1234567890,
                  "status": "abc123",
                  "redirectionURIs": "abc123",
                  "developerAppName": "abc123",
                  "developerId": "abc123" 
                }
              }`

		expectedResponse := fmt.Sprintf(template, headerValue)

		//post to the endpoint, then return a 401
		server.NewPost("/verifiers/apikey", "application/x-www-form-urlencoded", formBytes).ToResponse(http.StatusOK, []byte(expectedResponse)).Add()

		//start the server
		err := server.StartAsync("127.0.0.1:8181")
		defer server.Shutdown()

		Expect(err).Should(BeNil(), "Server should start")

		//create a mock reqeust.
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if err != nil {
			log.Fatal(err)
		}

		//add the api key
		req.Header.Add(headerKey, headerValue)

		w := httptest.NewRecorder()

		requestHandler(w, req)

		Expect(w.Code).Should(Equal(http.StatusOK), "Api Key")

		body := w.Body.String()

		Expect(body).Should(Equal(""), "Body")

	})
})

//mock endpoint url. "http://localhost:8181/verifiers/apikey"
func createPipeline() pipeline.Pipe {

	conf := make(map[interface{}]interface{})
	conf["apidUri"] = "http://localhost:8181/verifiers/apikey"
	conf["organization"] = organization
	conf["environment"] = environment
	conf["keyHeader"] = headerKey

	fitting, err := CreateFitting(conf)

	Expect(err).Should(BeNil(), "Error creating fitting")
	Expect(fitting).ShouldNot(BeNil(), "Error creating fitting")

	// //create our handler from our fitting
	handler := fitting.RequestHandlerFunc()

	reqFittings := []pipeline.FittingWithID{test_util.NewFittingFromHandlers("test", handler, nil)}

	pipeDef := pipeline.NewDefinition(reqFittings, []pipeline.FittingWithID{})

	pipe := pipeDef.CreatePipe(fmt.Sprintf("%d", time.Now().UnixNano()))

	return pipe
}

// http://localhost:8181/verifiers/apikey
