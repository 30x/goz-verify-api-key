package pipeline_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/30x/gozerian/pipeline"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Definition", func() {

	It("should load from a simple yaml struct", func() {

		// write a die & fitting (in another module)
		testDie := func(config interface{}) (Fitting, error) {
			fmt.Printf("test_fitting created: %v\n", config)
			return &test_fitting{config}, nil
		}

		// register the fitting die
		RegisterDies(map[string]Die{
			"test_die": testDie,
		})

		// define the pipe via YAML (loaded from URI)
		pipeYaml := `
request:
  - test_die:
      test_config: "3.14"
response:
  - test_die:
      test_config: "3.15"
`
		yamlReader := strings.NewReader(pipeYaml)

		// define a pipe
		pipeDef, err := DefinePipe(yamlReader)
		Expect(err).NotTo(HaveOccurred())

		pipe := pipeDef.CreatePipe("test_req_pipe")

		rec := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "URI", nil)
		Expect(err).NotTo(HaveOccurred())

		pipe.RequestHandlerFunc()(rec, req)
		b, err := ioutil.ReadAll(rec.Body)
		Expect(err).NotTo(HaveOccurred())

		val := string(b)
		Expect(val).To(Equal("3.14"))


		pipe = pipeDef.CreatePipe("test_res_pipe")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "URI", nil)
		Expect(err).NotTo(HaveOccurred())

		pipe.ResponseHandlerFunc()(rec, req, nil)
		b, err = ioutil.ReadAll(rec.Body)
		Expect(err).NotTo(HaveOccurred())

		val = string(b)
		Expect(val).To(Equal("3.15"))

	})
})

type test_fitting struct {
	config interface{}
}

func (f *test_fitting) getConfigVal(val string) interface{} {
	return f.config.(map[interface{}]interface{})[val]
}

func (f *test_fitting) RequestHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val := f.getConfigVal("test_config").(string)
		_, err := w.Write([]byte(val))
		Expect(err).NotTo(HaveOccurred())
	}
}

func (f *test_fitting) ResponseHandlerFunc() ResponseHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {
		val := f.getConfigVal("test_config").(string)
		_, err := w.Write([]byte(val))
		Expect(err).NotTo(HaveOccurred())
	}
}
