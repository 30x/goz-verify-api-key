package pipeline

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/context"
)

// ControlHolder holds a Control
type ControlHolder interface {
	Control() Control
}

// ExtraData is just an open map for storing stuff
type FlowData map[string]interface{}

// Control contains control and context for a Pipe instance
type Control interface {
	RequestID() string

	SetErrorHandler(eh ErrorHandlerFunc)
	ErrorHandler() ErrorHandlerFunc
	SendError(err interface{}) error
	Error() error

	Cancel()
	Cancelled() bool

	Log() Logger

	Config() config

	FlowData() FlowData

	Writer() http.ResponseWriter
}

type control struct {
	ctx          context.Context
	writer       http.ResponseWriter
	errorHandler ErrorHandlerFunc
	conf         config
	logger       Logger
	cancel       context.CancelFunc
	reqID        string
	flowData     FlowData
}

func (c *control) FlowData() FlowData {
	return c.flowData
}

func (c *control) RequestID() string {
	return c.reqID
}

func (c *control) Config() config {
	return c.conf
}

func (c *control) Log() Logger {
	return c.logger
}

func (c *control) ErrorHandler() ErrorHandlerFunc {
	if c.errorHandler == nil {
		return DefaultErrorHanderFunc
	}
	return c.errorHandler
}

func (c *control) SetErrorHandler(eh ErrorHandlerFunc) {
	c.Log().Debug("SetErrorHandler", eh)
	c.errorHandler = eh
}

func (c *control) SendError(r interface{}) error {
	if c.Cancelled() {
		return errors.New("Cancelled response, unable to send")
	}
	c.Log().Debug("SendError: ", r)
	var err error
	if reflect.TypeOf(r).String() != "error" {
		err = r.(error)
	} else {
		err = fmt.Errorf("%s", r)
	}
	return c.ErrorHandler()(c.writer, err)
}

func (c *control) Cancel() {
	c.Log().Debug("Cancel pipe")
	if c.Error() == nil {
		c.cancel()
	}
}

func (c *control) Error() error {
	return c.ctx.Err()
}

func (c *control) Cancelled() bool {
	return c.Error() != nil
}

func (c *control) Writer() http.ResponseWriter {
	return c.writer
}
