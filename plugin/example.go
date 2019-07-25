package main

import (
	"../cfdnsu"
	"encoding/json"
)

type s_configuration struct {
	Example struct {
		Hello string `json:"hello"`
	} `json:"example"`
}

var configuration s_configuration

func Startup() error {
	err := json.Unmarshal(cfdnsu.SharedInformation.Configuration, &configuration)

	if err != nil {
		cfdnsu.SharedInformation.Logger.Error(err)
		return err
	}

	if configuration.Example.Hello != "" {
		cfdnsu.SharedInformation.Logger.Infof("hello %s", configuration.Example.Hello)
	} else {
		cfdnsu.SharedInformation.Logger.Infof("looks like configuration.example.hello is missing check your .conf")
	}

	return nil
}

func Shutdown() error {
	cfdnsu.SharedInformation.Logger.Info("Shutdown issued!")

	return nil
}

func IpChanged() error {
	cfdnsu.SharedInformation.Logger.Infof("IpChanged to %s you are %c", cfdnsu.SharedInformation.CurrentIp, '\U0001F32F')

	return nil
}

func IpUpdated() error {
	cfdnsu.SharedInformation.Logger.Info("Ip record successfully updated!")

	return nil
}