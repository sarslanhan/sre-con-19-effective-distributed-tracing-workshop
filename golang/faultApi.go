package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	getFaultConfigsUrl = "https://sre-con-fault-injection-api.stups-test.zalan.do/api/faults"
)

type FaultConfig struct {
	LatencyMax   int `json:"latencyMax"`
	ErrorRateMax int `json:"errorRateMax"`
	Az           int `json:"az"`
}

type applicationFaultConfigTransfer struct {
	Configs map[string][]FaultConfig `json:"configs"`
}

type FaultInjectionAPI struct {
	client *http.Client
}

func NewFaultInjectionAPI() *FaultInjectionAPI {
	return &FaultInjectionAPI{client: http.DefaultClient}
}

func (api *FaultInjectionAPI) GetFaultConfigs() (map[string][]FaultConfig, error) {
	req, err := http.NewRequest(http.MethodGet, getFaultConfigsUrl, nil)
	if err != nil {
		return nil, err
	}

	res, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("Got HTTP response with status code %d ", res.StatusCode)
	}

	var responsePayload applicationFaultConfigTransfer
	err = json.NewDecoder(res.Body).Decode(&responsePayload)
	if err != nil {
		return nil, err
	}

	if len(responsePayload.Configs) == 0 {
		return nil, fmt.Errorf("HTTP Response from Fault Injection API came back empty")
	}

	return responsePayload.Configs, nil
}
