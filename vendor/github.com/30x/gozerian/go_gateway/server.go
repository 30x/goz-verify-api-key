package go_gateway

import (
	"github.com/30x/gozerian/pipeline"
	"io/ioutil"
	"bytes"
	"net/url"
	"net/http"
	"fmt"
	"io"
	"gopkg.in/yaml.v2"
)

/*
Note: Register Dies for pipeline before calling ListenAndServe.

Example of YAML:

port: 8080
target: http://httpbin.org
pipes:                      # pipe definitions
  main:                     # pipe id
    request:                # request pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
    response:               # response pipeline
    - dump:                 # name of plugin
        dumpBody: true      # plugin-specific configuration
proxies:                    # maps host & path -> pipe
  - host: localhost         # host
    path: /                 # path
    pipe: main              # pipe to use
 */
func ListenAndServe(yamlReader io.Reader) error {
	yamlIn, err := ioutil.ReadAll(yamlReader)
	if err != nil {
		return err
	}
	gateway := gatewayDef{}
	yaml.Unmarshal(yamlIn, &gateway)

	pipeDefinitions := make(map[string]pipeline.Definition)
	for id, def := range gateway.Pipes {
		pipeDefYaml, err := yaml.Marshal(def)
		if err != nil {
			return err
		}

		pipeDef, err := pipeline.DefinePipe(bytes.NewReader(pipeDefYaml))
		if err != nil {
			return err
		}

		pipeDefinitions[id] = pipeDef
	}

	target, err := url.Parse(gateway.Target)
	if err != nil {
		panic(err)
	}

	for _, proxy := range gateway.Proxies {
		proxyHandler := ReverseProxyHandler{pipeDefinitions[proxy.Pipe], target}
		http.HandleFunc(proxy.Path, proxyHandler.ServeHTTP)
	}

	fmt.Printf("Gateway routing port %s => %s\n", gateway.Port, target)
	err = http.ListenAndServe(":" + gateway.Port, nil)
	return err
}

type fittingDef map[string]interface{}

type pipeDef struct {
	Request  []fittingDef `json:"request"`
	Response []fittingDef `json:"response"`
}

type proxyDef struct {
	Host string `json:"host"`
	Path string `json:"path"`
	Pipe string `json:"pipe"`
}

type gatewayDef struct {
	Port   string             `json:"port"`
	Target string             `json:"target"`
	Pipes  map[string]pipeDef `json:"pipes"`
	Proxies []proxyDef        `json:"proxies"`
}
