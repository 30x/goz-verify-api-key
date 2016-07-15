package go_gateway

import (
	"net/http"
	"net"
	"bufio"
	"errors"
	"time"
	"io"
	"github.com/30x/gozerian/pipeline"
)

type targetTransport struct {
	http.RoundTripper
	control    pipeline.Control
	writer     http.ResponseWriter
	origReq	   *http.Request
	resHandler pipeline.ResponseHandlerFunc
}

func (tt *targetTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {

	// upgrade to hijacked connection
	upgrade := tt.origReq.Header.Get("Connection") == "Upgrade"
	if upgrade {
		return tt.upgradedRoundTrip(req)
	}

	// call target
	res, err = tt.RoundTripper.RoundTrip(req)
	if err != nil {
		tt.control.SendError(err)
	}

	// run response handlers
	if !tt.control.Cancelled() {
		tt.resHandler(tt.writer, tt.origReq, res)
	}

	return res, tt.control.Error()
}

func (self *targetTransport) upgradedRoundTrip(req *http.Request) (res *http.Response, err error) {

	req.Header.Set("Connection", self.origReq.Header.Get("Connection"))
	req.Header.Set("Upgrade", self.origReq.Header.Get("Upgrade"))

	targetConn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		return nil, err
	}
	defer targetConn.Close()

	if err := req.Write(targetConn); err != nil {
		return nil, err
	}

	targetReader := bufio.NewReader(targetConn)
	res, err = http.ReadResponse(targetReader, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 101 {
		return nil, errors.New("Upgrade status code (101) required")
	}

	// Run Response Handlers
	self.resHandler(self.writer, self.origReq, res)

	//log.Print("hijacking")
	con, clientRW, err := self.writer.(http.Hijacker).Hijack()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	res.Write(clientRW)
	clientRW.Flush()

	done := make(chan error)

	// set timeout timer
	// todo: can we get the elapsed time left instead?
	timeout := self.control.Config().GetDuration(pipeline.ConfigTimeout)
	timer := time.AfterFunc(timeout, func() { close(done) })

	// todo: pipe data through upgraded stream handlers!

	go copyData(clientRW, targetConn, done, timer, timeout)
	go copyData(targetConn, clientRW, done, timer, timeout)

	err = <-done

	// todo: add test for timeout

	return nil, err
}

func copyData(writer io.Writer, reader io.Reader, done chan error, timer *time.Timer, timeout time.Duration) {

	buf := make([]byte, 100, 100)
	c := time.Tick(time.Millisecond)
	for range c {
		n, err := io.CopyBuffer(writer, reader, buf)
		if err != nil {
			done <- err
			break
		}
		if n > 0 {
			timer.Reset(timeout)
		}
	}
}
