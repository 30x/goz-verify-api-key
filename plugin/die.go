/*
This example Gozerian plugin validates an API Key passed in a header by contacting the Apigee Edge Micro proxy on the server. If the API Key is missing the client receives a 401. If the API Key is incorrect, the client will receive a 403. If the API Key is valid, the key will be deleted from the header and the request will continue.

Sample Configuration:

	proxyURI:     https://sganyo-test.apigee.net/remoteproxy/v2/oauth/verifyApiKey
	proxyKey:     customerKeyForTheProxy
	apiKeyHeader: X-Apigee-API-Key
 */
package plugin

import (
	"fmt"
	"net/http"
	"github.com/30x/gozerian/pipeline"
	"encoding/json"
	"bytes"
	"io/ioutil"
	"errors"
	"regexp"
)

const VALID_URL string = `^((https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]+([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`


// exported function to create the fitting
func CreateFitting(config interface{}) (pipeline.Fitting, error) {

	conf, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("Invalid config. Expected map[interface{}]interface{}")
	}

	c := verifyAPIKeyConfig{
		conf["proxyURI"].(string),
		conf["proxyKey"].(string),
		conf["apiKeyHeader"].(string),
	}

	match, err := regexp.Match(VALID_URL, []byte(c.proxyURI))
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, fmt.Errorf("invalid proxyURI: %s", c.proxyURI)
	}

	if c.proxyKey == "" {
		return nil, fmt.Errorf("invalid config: missing proxyKey")
	}

	if c.apiKeyHeader == "" {
		return nil, fmt.Errorf("invalid config: missing apiKeyHeader")
	}

	return &verifyAPIKeyFitting{c}, nil
}

type (
	verifyAPIKeyConfig struct {
		proxyURI     string
		proxyKey     string
		apiKeyHeader string
	}

	ReqBody struct {
		ApiKey string `json:"apiKey"`
	}

	ResBody struct {
		Error string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
)


type verifyAPIKeyFitting struct {
	config verifyAPIKeyConfig
}

func (f *verifyAPIKeyFitting) RequestHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		control := w.(pipeline.ControlHolder).Control()
		log := control.Log()

		apiKey := r.Header.Get(f.config.apiKeyHeader)
		if apiKey == "" {
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
			return
		}

		reqBody := ReqBody{
			ApiKey: apiKey,
		}
		bytesBody, err := json.Marshal(reqBody)
		if err != nil {
			log.Debugf("unable to marshal request: %s\n", err.Error())
			control.SendError(err)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", f.config.proxyURI, bytes.NewReader(bytesBody))
		if err != nil {
			log.Debugf("error creating request: %s\n", err.Error())
			control.SendError(err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-DNA-Api-Key", f.config.proxyKey)
		res, err := client.Do(req)
		if err != nil {
			log.Debugf("error posting to server: %s\n", err.Error())
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

		if res.StatusCode < 200 || res.StatusCode >= 300 {
			log.Debugf("error from server: %s\n", string(resBytes))
			control.SendError(fmt.Errorf("error from server: %s (%d)", string(resBytes), res.StatusCode))
			return
		}

		result := ResBody{}
		err = json.Unmarshal(resBytes, &result)
		if err != nil {
			log.Debugf("error unmarshalling response: %s\n", string(resBytes))
			control.SendError(err)
			return
		}

		if result.Error != "" {
			w.WriteHeader(403)
			_, err = w.Write(resBytes)
			if err != nil {
				control.SendError(err)
			}
			return
		}

		r.Header.Del(f.config.apiKeyHeader)

		log.Debugln("Successful validation!")
	}
}

func (f *verifyAPIKeyFitting) ResponseHandlerFunc() pipeline.ResponseHandlerFunc {
	return nil
}
