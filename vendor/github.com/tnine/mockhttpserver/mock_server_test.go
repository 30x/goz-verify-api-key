package mockhttpserver_test

import (
	"io/ioutil"
	"time"

	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tnine/mockhttpserver"
)

var _ = Describe("Test Mock Server", func() {

	It("Simple GET ", func() {
		server := &MockServer{}

		reader := strings.NewReader("hello")

		server.NewGet("/foo").ToResponse(http.StatusOK, reader).Add()

		host := "127.0.0.1:5280"

		//start the server
		err := server.StartAsync(host)

		Expect(err).Should(BeNil(), "Server should start")

		//now hit it with http client
		resp, err := http.Get("http://127.0.0.1:5280/foo")

		Expect(err).Should(BeNil(), "Should get foo %s", err)

		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		bytes, err := ioutil.ReadAll(resp.Body)

		Expect(err).Should(BeNil(), "Should read response body")

		resp.Body.Close()

		stringVal := string(bytes)

		Expect(stringVal).Should(Equal("hello"), "Body should equal 'hello'")

		//now shut down the server. In a real test you would defer this immediately after start
		server.Shutdown()

		//try the request again, it shouldn't connect

		_, err = http.Get("http://127.0.0.1:5280/foo")

		Expect(err).ShouldNot(BeNil(), "Should not be able to do get")

	})

	It("Simple GET with headers", func() {
		server := &MockServer{}

		reader := strings.NewReader("hello")

		server.NewGet("/foo").AddHeader("TestHeader", "TestValue").ToResponse(http.StatusOK, reader).Add()

		host := "127.0.0.1:5280"

		//start the server
		err := server.StartAsync(host)

		Expect(err).Should(BeNil(), "Server should start")

		//now hit it with http client
		resp, err := http.Get("http://127.0.0.1:5280/foo")

		//should get a 404, since we didn't add the headers
		Expect(resp.StatusCode).Should(Equal(http.StatusNotFound))

		//now try again with the headers

		headerRequest, err := http.NewRequest("GET", "http://127.0.0.1:5280/foo", nil)

		headerRequest.Header.Set("TestHeader", "TestValue")

		client := &http.Client{
			Timeout: 120 * time.Second,
		}

		resp, err = client.Do(headerRequest)

		Expect(err).Should(BeNil(), "Should get foo %s", err)

		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		bytes, err := ioutil.ReadAll(resp.Body)

		Expect(err).Should(BeNil(), "Should read response body")

		resp.Body.Close()

		stringVal := string(bytes)

		Expect(stringVal).Should(Equal("hello"), "Body should equal 'hello'")

		//now shut down the server. In a real test you would defer this immediately after start
		server.Shutdown()

		//try the request again, it shouldn't connect

		_, err = http.Get("http://127.0.0.1:5280/foo")

		Expect(err).ShouldNot(BeNil(), "Should not be able to do get")

	})
})
