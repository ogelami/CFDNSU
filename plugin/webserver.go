package main

import (
	"fmt"
	"../cfdnsu"
)

func Startup() {
	cfdnsu.SharedInformation.Logger.Infof("tototoot")
	fmt.Printf(cfdnsu.SharedInformation.CurrentIp)
}

func Shutdown() {
	fmt.Printf("Shutdown")
}

func IpChanged() {
	fmt.Printf("IpChanged to %s", cfdnsu.SharedInformation.CurrentIp)
}