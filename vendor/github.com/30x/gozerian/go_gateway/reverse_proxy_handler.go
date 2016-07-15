package go_gateway

// This is used for a standalone Go gateway utilizing http.ReverseProxy.
// It is used for unit tests and such, but it's not expected to be a production implementation.

import (
	"net/url"
	"net/http"
	"path"
	"net/http/httputil"
	"github.com/30x/gozerian/pipeline"
	"net"
	"bufio"
	"log"
	"os"
)

type ReverseProxyHandler struct {
	Definition pipeline.Definition
	Target     *url.URL
}

func (self *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	req.URL.Host = self.Target.Host

	// todo: Set up pipes, select pipe by host & path
	pipe := self.Definition.CreatePipe("") // leaving blank to auto-assign ID

	// call request handlers
	pipe.RequestHandlerFunc()(w, req)

	// abort pipeline as needed
	if pipe.Control().Cancelled() {
		return
	}

	// call target & response handlers
	targetQuery := self.Target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = self.Target.Scheme
		req.URL.Path = path.Join(self.Target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	resHandler := pipe.ResponseHandlerFunc()

	// if I return an error, RoundTripper attempts to write a 500
	// wrap my writer to avoid that WriteHeader call, since I deal with it elsewhere
	// also provide a logger that avoids extraneous log messages
	ch := pipe.Control()
	ww := &resWriter{ch.Writer(), ch, w}
	log := log.New(&logWriter{ch}, "", 0)

	transport := &targetTransport{http.DefaultTransport, pipe.Control(), ww, req, resHandler}
	targetProxy := &httputil.ReverseProxy{Director: director, Transport: transport, ErrorLog: log}
	targetProxy.ServeHTTP(ww, req)
}

type logWriter struct {
	control pipeline.Control
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	if !lw.control.Cancelled() {
		return os.Stderr.Write(p)
	}
	return 0, nil
}

type resWriter struct {
	writer http.ResponseWriter
	control pipeline.Control
	hijackWriter http.ResponseWriter
}

func (w *resWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *resWriter) Write(bytes []byte) (int, error) {
	return w.writer.Write(bytes)
}

func (w *resWriter) WriteHeader(status int) {
	if w.control.Cancelled() {
		return
	}
	w.writer.WriteHeader(status)
}

func (w *resWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.hijackWriter.(http.Hijacker).Hijack()
}