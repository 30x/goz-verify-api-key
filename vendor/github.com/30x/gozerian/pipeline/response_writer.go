package pipeline

import (
	"net/http"
)

func newResponseWriter(writer http.ResponseWriter, control Control) responseWriter {
	return &resWriter{writer, control}
}

type responseWriter interface {
	http.ResponseWriter
	ControlHolder
}

type resWriter struct {
	writer  http.ResponseWriter
	control Control
}

func (w *resWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *resWriter) Write(bytes []byte) (int, error) {
	w.control.Log().Debugf("Write: %s", string(bytes))
	w.control.Cancel()
	return w.writer.Write(bytes)
}

func (w *resWriter) WriteHeader(status int) {
	w.control.Log().Debugf("WriteHeader: %d", status)
	w.control.Cancel()
	w.writer.WriteHeader(status)
}

func (w *resWriter) Control() Control {
	return w.control
}
