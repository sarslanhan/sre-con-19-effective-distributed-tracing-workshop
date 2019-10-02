package main

import (
	"log"
	"os"
	"strconv"
)

const (
	azEnvVarName         = "AVAILABILITY_ZONE"
	instanceIDEnvVarName = "INSTANCE_ID"
)

func main() {
	azEnvVarValue := os.Getenv(azEnvVarName)
	azProperty, err := strconv.Atoi(azEnvVarValue)
	if err != nil {
		log.Fatalf("Error converting AZ Env Var (named %s) (current value: %s)", azEnvVarName, azEnvVarValue)
	}

	instanceID := os.Getenv(instanceIDEnvVarName)
	if instanceID == "" {
		log.Fatalf("Please use %s environment variable to name your instance. It can be anything you want.", instanceIDEnvVarName)
	}

	faultManager := NewFaultInjectionManager(azProperty)
	faultManager.Run()

	website := newWebsite(faultManager, instanceID, azProperty)
	cartAPI := newCartAPI(faultManager, instanceID, azProperty)

	go func() { log.Fatal(cartAPI.Run(":8085")) }()
	log.Fatal(website.Run(":8087"))
}
