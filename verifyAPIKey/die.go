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
package verifyAPIKey

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/30x/gozerian/pipeline"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"github.com/afex/hystrix-go/hystrix"
)

const DEFAULT_APID_URI = "http://localhost:8181/verifiers/apikey"
const DEFAULT_KEY_HEADER = "X-Apigee-API-Key"
const UNAUTHORIZED_MESSAGE = "Unauthorized"

type verifyAPIKeyConfig struct {
	apidUri      string
	keyHeader    string
	organization string
	environment  string
	send404OnError bool
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

	// validate all config entries

	if conf["organization"] == nil {
		return nil, fmt.Errorf("invalid config: missing organization")
	}

	if conf["environment"] == nil {
		return nil, fmt.Errorf("invalid config: missing environment")
	}

	if conf["apidUri"] == nil {
		conf["apidUri"] = DEFAULT_APID_URI
	}

	if conf["keyHeader"] == nil {
		conf["keyHeader"] = DEFAULT_KEY_HEADER
	}

	if conf["send404OnError"] == nil {
		conf["send404OnError"] = false
	}

	c := verifyAPIKeyConfig{
		apidUri:        conf["apidUri"].(string),
		keyHeader:      conf["keyHeader"].(string),
		organization:   conf["organization"].(string),
		environment:    conf["environment"].(string),
		send404OnError: conf["send404OnError"].(bool),
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
			if f.config.send404OnError {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(UNAUTHORIZED_MESSAGE))
			}
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

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		err = hystrix.Do(DEFAULT_APID_URI, func() error {

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				log.Debugf("error getting from server: %s\n", err.Error())
				control.SendError(err)
				return nil
			}
			defer res.Body.Close()

			resBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Debugf("error reading reply: %s\n", err.Error())
				control.SendError(err)
				return nil
			}

			// if we get an error, log and send 500 to client
			if res.StatusCode < 200 || res.StatusCode >= 300 {
				if f.config.send404OnError {
					w.WriteHeader(http.StatusNotFound)
				} else {
					log.Debugf("key verification code: %d message: %s\n", res.StatusCode, string(resBytes))
					control.SendError(fmt.Errorf("internal error with key verification"))
				}
				return nil
			}

			// check for failed
			failResponse := KMSResponseFail{}
			err = json.Unmarshal(resBytes, &failResponse)
			if err != nil {
				control.SendError(err)
				return nil
			}
			if failResponse.Errinfo.ErrorCode != "" {
				log.Debugf("key verification failed: %s\n", failResponse.Errinfo)
				if f.config.send404OnError {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(UNAUTHORIZED_MESSAGE))
				}
				return nil
			}

			// store valid response in flowdata
			response := KMSResponseSuccess{}
			err = json.Unmarshal(resBytes, &response)
			if err != nil {
				control.SendError(err)
				return nil
			}
			flowData := control.FlowData()
			for k, v := range response.Rspinfo {
				flowData[k] = v
				log.Debugf("Setting flow var: %s to: %v\n", k, v)
			}

			log.Debugln("Successful validation!")

			return nil
		}, nil)

		if err != nil {
			control.SendError(err)
		}
	}
}

func (f *verifyAPIKeyFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return nil
}
