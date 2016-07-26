package c_gateway_test

import (
	. "github.com/30x/gozerian/pipeline"
	"github.com/30x/gozerian/test_util"
	"net/http/httptest"
	"net/url"
	"github.com/30x/gozerian/go_gateway"
	"net/http"
)

// Test framework: http://onsi.github.io/ginkgo/

func init() {
	RegisterDie("dump", test_util.CreateDumpFitting)
}

func makeGateway(targetURL string, reqHands []http.HandlerFunc, resHands []ResponseHandlerFunc) *httptest.Server {

	dumpConfig := make(map[interface{}]interface{})
	dumpConfig["dumpBody"] = false
	dumpFitting, err := NewFitting("dump", dumpConfig)
	if err != nil {
		panic(err)
	}

	var reqFittings []FittingWithID
	reqFittings = append(reqFittings, dumpFitting)
	for _, h := range reqHands {
		reqFittings = append(reqFittings, test_util.NewFittingFromHandlers("test", h, nil))
	}

	var resFittings []FittingWithID
	resFittings = append(resFittings, dumpFitting)
	for _, h := range resHands {
		resFittings = append(resFittings, test_util.NewFittingFromHandlers("test", nil, h))
	}

	pipeDef := NewDefinition(reqFittings, resFittings)
	if err != nil {
		panic(err)
	}
	target, _:= url.Parse(targetURL)
	proxyHandler := &go_gateway.ReverseProxyHandler{pipeDef, target}

	return httptest.NewServer(proxyHandler)
}

var _ = test_util.TestPipelineAgainst(makeGateway)
