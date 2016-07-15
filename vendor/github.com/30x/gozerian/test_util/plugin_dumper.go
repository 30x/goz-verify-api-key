package test_util

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"errors"
	"net/http/httputil"
	"io"
)

// export function to create the fitting
func CreateDumpFitting(config interface{}) (pipeline.Fitting, error) {

	conf, ok := config.(map[interface{}]interface{})
	if (!ok) {
		return nil, errors.New("Invalid config. Expected map[interface{}]interface{}")
	}
	c := dumpFittingConfig{
		conf["dumpBody"].(bool),
	}

	return &dumpFitting{c}, nil
}

type dumpFittingConfig struct {
	dumpBody bool
}


type dumpFitting struct {
	config dumpFittingConfig
}

func (f *dumpFitting) RequestHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		control := w.(pipeline.ControlHolder).Control()
		log := control.Log()
		if f.config.dumpBody {
			r.Body = loggingReadCloser{r.Body, log, "body >>"}
		}
		log.Println("======================== request ========================")
		dump, err := httputil.DumpRequest(r, false)
		if err != nil {
			w.(pipeline.ControlHolder).Control().SendError(err)
		}
		log.Print(string(dump))
	}
}

func (f *dumpFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, res *http.Response) {
		control := w.(pipeline.ControlHolder).Control()
		log := control.Log()
		if f.config.dumpBody {
			res.Body = loggingReadCloser{res.Body, log, "body <<"}
		}
		log.Println("======================== response ========================")
		dump, err := httputil.DumpResponse(res, false)
		if err != nil {
			w.(pipeline.ControlHolder).Control().SendError(err)
		}
		log.Print(string(dump))
	}
}

type loggingReadCloser struct {
	io.ReadCloser
	log       pipeline.Logger
	indicator string
}

func (l loggingReadCloser) Read(buf []byte) (n int, err error) {

	n, err = l.ReadCloser.Read(buf)
	if n > 0 {
		l.log.Printf("%s %q\n", l.indicator, string(buf[:n]))
	}
	return n, err
}
