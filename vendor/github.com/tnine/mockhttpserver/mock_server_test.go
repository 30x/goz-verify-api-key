package mockhttpserver_test

import (
	"io/ioutil"
	"time"

	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tnine/mockhttpserver"
)

var _ = Describe("Test Mock Server", func() {

	It("Simple GET ", func() {
		server := &MockServer{}

		hello := []byte("hello")

		server.NewGet("/foo").ToResponse(http.StatusOK, hello).Add()

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

		hello := []byte("hello")

		server.NewGet("/foo").AddHeader("TestHeader", "TestValue").ToResponse(http.StatusOK, hello).Add()

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

	It("Simple POST", func() {
		server := &MockServer{}

		hello := []byte("hello")

		server.NewPost("/foo", "text/plain", hello).ToResponse(http.StatusOK, hello).Add()

		host := "127.0.0.1:5280"

		//start the server
		err := server.StartAsync(host)

		Expect(err).Should(BeNil(), "Server should start")

		//now hit it with http client
		resp, err := http.Post("http://127.0.0.1:5280/foo", "text/plain", strings.NewReader("hello"))

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

	It("Form body  POST", func() {
		server := &MockServer{}

		form := url.Values{}
		form.Add("key1", "value1")
		form.Add("key2", "value2")

		formAsString := form.Encode()

		body := []byte(formAsString)

		accepted := []byte("accepted")

		server.NewPost("/foo", "application/x-www-form-urlencoded", body).ToResponse(http.StatusOK, accepted).Add()

		host := "127.0.0.1:5280"

		//start the server
		err := server.StartAsync(host)

		Expect(err).Should(BeNil(), "Server should start")

		//now hit it with http client
		resp, err := http.Post("http://127.0.0.1:5280/foo", "application/x-www-form-urlencoded", strings.NewReader(formAsString))

		Expect(err).Should(BeNil(), "Should get foo %s", err)

		Expect(resp.StatusCode).Should(Equal(http.StatusOK))

		bytes, err := ioutil.ReadAll(resp.Body)

		Expect(err).Should(BeNil(), "Should read response body")

		resp.Body.Close()

		stringVal := string(bytes)

		Expect(stringVal).Should(Equal("accepted"), "Body should equal 'accepted'")

		//now shut down the server. In a real test you would defer this immediately after start
		server.Shutdown()

		//try the request again, it shouldn't connect

		_, err = http.Get("http://127.0.0.1:5280/foo")

		Expect(err).ShouldNot(BeNil(), "Should not be able to do get")

	})
})
