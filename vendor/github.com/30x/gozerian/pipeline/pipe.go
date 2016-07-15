package pipeline

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Pipe runs a series of request and response handlers - one created per request
type Pipe interface {
	ControlHolder
	RequestHandlerFunc() http.HandlerFunc
	ResponseHandlerFunc() ResponseHandlerFunc
}

var reqCounter int64

func newPipe(reqID string, reqFittings []FittingWithID, resFittings []FittingWithID) Pipe {

	// provide a default (transient) implementation of an ID
	if reqID == "" {
		reqID = string(strconv.FormatInt(atomic.AddInt64(&reqCounter, 1), 10))
	}

	return &pipe{
		reqID:    reqID,
		reqFittings: reqFittings,
		resFittings: resFittings,
	}
}

type pipe struct {
	reqID       string
	reqFittings []FittingWithID
	resFittings []FittingWithID
	ctrl        *control
}

func (p *pipe) Control() Control {
	return p.ctrl
}

func (p *pipe) RequestHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		writer := p.setWriter(w, r)
		defer recoveryFunc(p.ctrl)

		for _, fitting := range p.reqFittings {
			if p.ctrl.Cancelled() {
				break
			}
			p.ctrl.Log().Debugf("Enter req handler %s", fitting.ID())
			fitting.RequestHandlerFunc()(writer, r)
			p.ctrl.Log().Debugf("Exit req handler %s", fitting.ID())
		}
	}
}

func (p *pipe) ResponseHandlerFunc() ResponseHandlerFunc {

	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {

		writer := p.setWriter(w, r)
		defer recoveryFunc(p.ctrl)

		for _, fitting := range p.resFittings {
			if p.ctrl.Cancelled() {
				break
			}
			p.ctrl.Log().Debugf("Enter res handler %s", fitting.ID())
			fitting.ResponseHandlerFunc()(writer, r, res)
			p.ctrl.Log().Debugf("Exit res handler %s", fitting.ID())
		}
	}
}

func (p *pipe) setWriter(w http.ResponseWriter, r *http.Request) http.ResponseWriter {

	if p.ctrl != nil && p.ctrl.Writer() != nil {
		return p.ctrl.Writer()
	}

	writer, ok := w.(responseWriter)
	if !ok {
		f := logrus.Fields{
			"reqID":  p.reqID,
			"uri": r.RequestURI,
		}
		log := getLogger().WithFields(f)

		ctx, cancel := context.WithTimeout(context.Background(), getConfig().GetDuration(ConfigTimeout))
		ctl := &control{
			reqID: p.reqID,
			ctx: ctx,
			conf: conf,
			logger: log,
			cancel: cancel,
			flowData: make(map[string]interface{}),
		}

		// todo: this is a weird do-si-do circular reference. clean up?
		writer = newResponseWriter(w, ctl)
		ctl.writer = writer

		p.ctrl = ctl
	}
	p.ctrl.writer = writer
	return writer
}

func (p *pipe) Writer() http.ResponseWriter {
	return p.ctrl.Writer()
}

func recoveryFunc(pc Control) {
	if r := recover(); r != nil {
		err := fmt.Errorf("%s", r)
		pc.Log().Warn("Panic Recovery Error: ", err)
		pc.SendError(err)
	}
}
