package test_util

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
)

// NewFittingFromHandlers creates a fitting from a request and response handler
func NewFittingFromHandlers(id string, reqHandler http.HandlerFunc, resHandler pipeline.ResponseHandlerFunc) pipeline.FittingWithID {
	return &handlerFitting{id, reqHandler, resHandler}
}

type handlerFitting struct {
	id         string
	reqHandler http.HandlerFunc
	resHandler pipeline.ResponseHandlerFunc
}

func (f *handlerFitting) ID() string {
	return f.id
}

func (f *handlerFitting) RequestHandlerFunc() http.HandlerFunc {
	return f.reqHandler
}

func (f *handlerFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return f.resHandler
}
