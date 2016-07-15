package verifyAPIKey

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/30x/gozerian/pipeline"
	"github.com/30x/gozerian/test_util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Mock Server", func() {

	It("No api key ", func() {
		//use the pipe fitting to create the pipeline
		pipeline := createPipeline()

		//get our request handler
		requestHandler := pipeline.RequestHandlerFunc()

		//create a mock reqeust.  Deliberately has no header
		req, err := http.NewRequest("GET", "http://example.com/foo", nil)
		if err != nil {
			log.Fatal(err)
		}

		w := httptest.NewRecorder()

		requestHandler(w, req)

		Expect(w.Code).Should(Equal(http.StatusUnauthorized), "No api key was presented")

	})
})

//mock endpoint url. "http://localhost:8181/verifiers/apikey"
func createPipeline() pipeline.Pipe {

	conf := make(map[interface{}]interface{})
	conf["apidUri"] = "http://localhost:8181/verifiers/apikey"
	conf["organization"] = "radical-new"
	conf["environment"] = "test"
	conf["keyHeader"] = "X-Apigee-API-Key"

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
