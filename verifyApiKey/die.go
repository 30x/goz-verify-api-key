/*
A Gozerian plugin that validates an API Key passed in a header by contacting the localhost apid server. If the API Key
is valid, the key will be deleted from the header, the data from the server will be stored in the control.FlowData and
the request will continue.

Sample Configuration:

    - verifyAPIKey:
        apidUri: http://localhost:8181/verifiers/apikey
        organization: radical-new
        environment:  test
        keyHeader: X-Apigee-API-Key
*/
package verifyApiKey

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/30x/gozerian/pipeline"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const DEFAULT_APID_URI = "http://localhost:8181/verifiers/apikey"
const DEFAULT_KEY_HEADER = "X-Apigee-API-Key"

type verifyAPIKeyConfig struct {
	apidUri      string
	keyHeader    string
	organization string
	environment  string
}

type verifyAPIKeyFitting struct {
	config verifyAPIKeyConfig
}

type ErrResponseDetail struct {
	ErrorCode string `json:"errorCode"`
	Reason    string `json:"reason"`
}

type KMSResponseSuccess struct {
	Rspinfo map[string]interface{} `json:"result"`
	ResponseType string            `json:"responseType"`
	ResponseCode int               `json:"responseCode"`
}

type KMSResponseFail struct {
	Errinfo      ErrResponseDetail `json:"result"`
	ResponseType string            `json:"responseType"`
	ResponseCode int               `json:"responseCode"`
}


// CreateFitting exported function to create the fitting
func CreateFitting(config interface{}) (pipeline.Fitting, error) {

	conf, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("Invalid config. Expected map[interface{}]interface{}")
	}

	c := verifyAPIKeyConfig{
		apidUri:      conf["apidUri"].(string),
		keyHeader:    conf["keyHeader"].(string),
		organization: conf["organization"].(string),
		environment:  conf["environment"].(string),
	}

	if c.apidUri == "" {
		c.apidUri = DEFAULT_APID_URI
	}

	if c.keyHeader == "" {
		c.keyHeader = DEFAULT_KEY_HEADER
	}

	if c.organization == "" {
		return nil, fmt.Errorf("invalid config: missing organization")
	}

	if c.environment == "" {
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

		form := url.Values{}
		form.Add("action", "verify")
		form.Add("organization", f.config.organization)
		form.Add("environment", f.config.environment)
		form.Add("key", apiKey)
		form.Add("uriPath", r.URL.Path)
		log.Debugf("Posting %s with body:\n", f.config.apidUri, form.Encode())

		req, err := http.NewRequest("POST", f.config.apidUri, strings.NewReader(form.Encode()))
		if err != nil {
			log.Debugf("error creating request: %s\n", err.Error())
			control.SendError(err)
			return
		}

		//req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
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
			log.Debugf("key verification error: %s\n", string(resBytes))
			control.SendError(fmt.Errorf("key verification error: %s (%d)", string(resBytes), res.StatusCode))
			return
		}

		// check for failed
		failResponse := KMSResponseFail{}
		err = json.Unmarshal(resBytes, &failResponse)
		if err != nil {
			control.SendError(err)
			return
		}
		if failResponse.Errinfo.ErrorCode != "" {
			msg, _ := json.Marshal(failResponse.Errinfo)
			// TODO: this needs to be revisited!!!
			// 1. the response code should not be dictated by apid
			// 2. should it return http.StatusUnauthorized? or 404? should this be an option?
			respCode := failResponse.ResponseCode
			if respCode == 404 {
				respCode = http.StatusUnauthorized
			}
			w.WriteHeader(respCode)
			w.Write([]byte(msg))
			return
		}

		// store valid response in flowdata
		response := KMSResponseSuccess{}
		err = json.Unmarshal(resBytes, &response)
		if err != nil {
			control.SendError(err)
			return
		}
		flowData := control.FlowData()
		for k, v := range response.Rspinfo {
			flowData[k] = v
			log.Debugf("Setting flow var: %s to: %v\n", k, v)
		}

		log.Debugln("Successful validation!")
	}
}

func (f *verifyAPIKeyFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return nil
}
