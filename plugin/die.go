/*
This example Gozerian plugin validates an API Key passed in a header by contacting the Apigee Edge Micro proxy on the server. If the API Key is missing the client receives a 401. If the API Key is incorrect, the client will receive a 403. If the API Key is valid, the key will be deleted from the header, the data from the server will be stored in the control.FlowData and the request will continue.

Sample Configuration:

	proxyURI:     https://sganyo-test.apigee.net/remoteproxy/v2/oauth/verifyApiKey
	proxyKey:     customerKeyForTheProxy
	apiKeyHeader: X-Apigee-API-Key
*/
package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/30x/gozerian/pipeline"
)

const VALID_URL string = `^((https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]+([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`

//CreateFitting exported function to create the fitting
func CreateFitting(config interface{}) (pipeline.Fitting, error) {

	conf, ok := config.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("Invalid config. Expected map[interface{}]interface{}")
	}

	c := verifyAPIKeyConfig{
		apidURL:      conf["apidURL"].(string),
		apiKeyHeader: conf["apiKeyHeader"].(string),
		apiOrg:       conf["apiOrg"].(string),
	}

	match, err := regexp.Match(VALID_URL, []byte(c.apidURL))
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, fmt.Errorf("invalid proxyURI: %s", c.apidURL)
	}

	if c.apiKeyHeader == "" {
		return nil, fmt.Errorf("invalid config: missing apiKeyHeader")
	}

	if c.apiOrg == "" {
		return nil, fmt.Errorf("invalid config: missing apiOrg")
	}

	return &verifyAPIKeyFitting{c}, nil
}

type verifyAPIKeyConfig struct {
	apidURL      string
	apiKeyHeader string
	apiOrg       string
}

type verifyAPIKeyFitting struct {
	config verifyAPIKeyConfig
}

func (f *verifyAPIKeyFitting) RequestHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		control := w.(pipeline.ControlHolder).Control()
		log := control.Log()

		apiKey := r.Header.Get(f.config.apiKeyHeader)

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
		req, err := http.NewRequest("GET", f.config.apidURL, nil)
		if err != nil {
			log.Debugf("error creating request: %s\n", err.Error())
			control.SendError(err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("apiKey", apiKey)
		req.Header.Set("org", f.config.apiOrg)
		req.Header.Set("path", path)

		log.Debugf("Requesting %s with apiKey:%s org:%s path:%s \n", req.URL.String(), apiKey, f.config.apiOrg, path)

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

		//if we get a forbidden, pass it along.  Any other error will be sent as well
		if res.StatusCode < 200 || res.StatusCode >= 300 {
			log.Debugf("error from server on key verification: %s\n", string(resBytes))
			control.SendError(fmt.Errorf("error from server on key verification: %s (%d)", string(resBytes), res.StatusCode))
			return
		}

		//TODO, do we want to delete the header?
		r.Header.Del(f.config.apiKeyHeader)

		//Take everything from the valid resposne and pass it along.
		//TODO, should the keys be namespaced so it's clear what plugin has populated them?
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

/* This is what a valid result from validate looks like:
{"developer_email":"remote-proxy@apigee.com","issued_at":1446155872425,"status":"approved","apiproduct_name":"remoteproxy","developer_app_id":"12a6600f-18ef-46a2-bd50-d3b529d18d68","expires_in":0,"client_id":"qFj141vqfH3dGyYv6useAmsxYWKeuCtO","developer_id":"sganyo@@@80c9db9b-8e2d-463f-90f4-988a8ad865f9","developer_app_name":"remoteproxy","attributes":{"client_secret":"YyAcpkSDHH7X8jXj","redirection_uris":null},"developer":{"apps":["remoteproxy"],"app_name":"remoteproxy","app_id":"12a6600f-18ef-46a2-bd50-d3b529d18d68","id":"sganyo@@@80c9db9b-8e2d-463f-90f4-988a8ad865f9","attributes":{"created_by":"sganyo@apigee.com","lastName":"Proxy","last_modified_by":"sganyo@apigee.com","last_modified_at":"1421284658925","status":"active","email":"remote-proxy@apigee.com","created_at":"1421284658925","userName":"remote-proxy","firstName":"Remote"}},"app":{"apiproducts":["remoteproxy"],"scopes":[],"attributes":{"created_by":"sganyo@apigee.com","id":"12a6600f-18ef-46a2-bd50-d3b529d18d68","callbackUrl":"","last_modified_by":"sganyo@apigee.com","last_modified_at":"1446155872242","status":"approved","appParentId":"sganyo@@@80c9db9b-8e2d-463f-90f4-988a8ad865f9","appParentStatus":"active","name":"remoteproxy","created_at":"1446155872242","appFamily":"default","appType":"Developer","accessType":null}}}
*/
