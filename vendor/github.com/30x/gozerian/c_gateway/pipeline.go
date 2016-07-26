package c_gateway

import (
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"net/url"
	"errors"
	"os"
	"io"
)

/*
Note: Register Dies for pipeline before calling ListenAndServe.

Example of expected YAML at URL:

    request:                # request pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
    response:               # response pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
 */

func DefinePipe(configURL *url.URL) (pipeline.Definition, error) {

	var reader io.Reader

	if (configURL.Scheme == "file") {
		file, err := os.Open(configURL.Path)
		if err != nil {
			return nil, err
		}
		reader = file
	} else {
		res, err := http.Get(configURL.String())
		if err != nil {
			return nil, err
		}
		if res.Body == nil {
			return nil, errors.New("Invalid URL, no body")
		}
		defer res.Body.Close()
		reader = res.Body
	}

	return pipeline.DefinePipe(reader)
}
