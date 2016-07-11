/*
A Gozerian plugin that validates an API Key passed in a header by contacting the localhost apid server. If the API Key
is valid, the key will be deleted from the header, the data from the server will be stored in the control.FlowData and
the request will continue.

Sample Configuration:

	keyHeader: X-Apigee-API-Key
	organization: test
	environment: test
*/
package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/30x/gozerian/pipeline"
	"io/ioutil"
	"net/http"
)

const APID_URI = "http://localhost:8081/verifiers/apikey"

type verifyAPIKeyConfig struct {
	keyHeader    string
	organization string
	environment  string
}

type verifyAPIKeyFitting struct {
	config verifyAPIKeyConfig
}

type verifyAPIBody struct {
	action       string
	organization string
	environment  string
	key          string
	uriPath      string
}

// CreateFitting exported function to create the fitting
func CreateFitting(config interface{}) (pipeline.Fitting, error) {

	conf, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("Invalid config. Expected map[interface{}]interface{}")
	}

	c := verifyAPIKeyConfig{
		keyHeader:    conf["keyHeader"].(string),
		organization: conf["organization"].(string),
		environment:  conf["environment"].(string),
	}

	if c.keyHeader == "" {
		return nil, fmt.Errorf("invalid config: missing keyHeader")
	}

	if c.organization == "" {
		return nil, fmt.Errorf("invalid config: missing organization")
	}

	if c.organization == "" {
		return nil, fmt.Errorf("invalid config: missing environment")
	}

	return &verifyAPIKeyFitting{c}, nil
}

func (f *verifyAPIKeyFitting) RequestHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		control := w.(pipeline.ControlHolder).Control()
		log := control.Log()

		apiKey := r.Header.Get(f.config.keyHeader)

		log.Debugf("Api config is %v.  Header value is %s. \n", f.config, apiKey)

		if apiKey == "" {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
			return
		}

		//parse the path
		path := r.URL.Path

		//create our request
		client := &http.Client{}

		msg := verifyAPIBody{
			action: "verify",
			"organization", f.config.organization,
			"environment", f.config.environment,
			"key", apiKey,
			"uriPath": path,
		}
		body, err := json.Marshal(msg)
		if err != nil {
			log.Debugf("error creating request body: %s\n", err.Error())
			control.SendError(err)
			return
		}

		req, err := http.NewRequest("POST", APID_URI, bytes.NewBuffer(body))
		if err != nil {
			log.Debugf("error creating request: %s\n", err.Error())
			control.SendError(err)
			return
		}

		log.Debugf("Posting %s with body:\n", req.URL.String(), body)

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Debugf("error getting from server: %s\n", err.Error())
			control.SendError(err)
			return
		}

		defer res.Body.Close()

		resBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Debugf("error reading reply: %s\n", err.Error())
			control.SendError(err)
			return
		}

		// if we get a forbidden, pass it along.  Any other error will be sent as well
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			log.Debugf("error from server on key verification: %s\n", string(resBytes))
			control.SendError(fmt.Errorf("error from server on key verification: %s (%d)", string(resBytes), res.StatusCode))
			return
		}

		// Take everything from the valid response and pass it along.
		newFlowVars := make(map[string]interface{})
		err = json.Unmarshal(resBytes, &newFlowVars)
		if err != nil {
			control.SendError(err)
			return
		}

		flowData := control.FlowData()
		for k, v := range newFlowVars {
			flowData[k] = v
		}

		log.Debugln("Successful validation!")
	}
}

func (f *verifyAPIKeyFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return nil
}
