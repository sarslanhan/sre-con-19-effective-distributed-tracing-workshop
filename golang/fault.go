package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	defaultLatencyMax   = 20
	defaultErrorRateMax = 0
	pollingInterval     = 2 * time.Second
)

// FaultInjectionManager responsible for produce unexpected problems or errors to the apps
type FaultInjectionManager struct {
	config          map[string]FaultConfig
	az              int
	api             *FaultInjectionAPI
	pollingInterval time.Duration
}

// NewFaultInjectionManager initialize FaultInjectionManager
func NewFaultInjectionManager(az int) *FaultInjectionManager {
	pollingIntervalProperty := pollingInterval

	defaultConfig := make(map[string]FaultConfig, 2)
	defaultConfig["cart-api"] = FaultConfig{
		LatencyMax:   defaultLatencyMax,
		ErrorRateMax: defaultErrorRateMax,
	}
	defaultConfig["the-super-site"] = FaultConfig{
		LatencyMax:   defaultLatencyMax,
		ErrorRateMax: defaultErrorRateMax,
	}

	return &FaultInjectionManager{
		config:          defaultConfig,
		az:              az,
		api:             NewFaultInjectionAPI(),
		pollingInterval: pollingIntervalProperty,
	}

}

func (manager *FaultInjectionManager) Run() {

	manager.updateConfigs()

	go func(manager *FaultInjectionManager) {
		for {
			time.Sleep(time.Duration(manager.pollingInterval))
			manager.updateConfigs()
		}
	}(manager)
}

func (manager *FaultInjectionManager) updateConfigs() {
	apiResponse, err := manager.api.GetFaultConfigs()
	if err != nil {
		log.Printf("Could not update configs due to error: %v", err)
		return
	}

	newFaultConfigs := make(map[string]FaultConfig, 2)
	for _, appName := range []string{siteAppName, cartAppName} {
		faultConfig, err := selectFaultConfig(apiResponse, appName, manager.az)
		if err != nil {
			log.Printf("Could not update configs. (reason: %v)", err)
			return
		}
		newFaultConfigs[appName] = *faultConfig
	}

	manager.config = newFaultConfigs
}

func selectFaultConfig(apiResponse map[string][]FaultConfig, appName string, az int) (*FaultConfig, error) {
	faultConfigList, exists := apiResponse[appName]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}
	for _, config := range faultConfigList {
		if config.Az == az {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("configuration for the AZ not found")
}

func (manager *FaultInjectionManager) sleepForAWhile(appName string) {
	maxSleep := manager.config[appName].LatencyMax
	r := 0
	if maxSleep > 0 {
		r = rand.Intn(maxSleep)
	}
	time.Sleep(10 + time.Duration(r)*time.Millisecond)

}

func shouldFail(errorRate int) bool {
	r := rand.Intn(100) - errorRate
	return r < 0
}

func (manager *FaultInjectionManager) maybeFailTheOperation(appName string) error {
	errorRate := manager.config[appName].ErrorRateMax
	if shouldFail(errorRate) {
		return fmt.Errorf("failed to add record to remote storage")
	}
	return nil
}
