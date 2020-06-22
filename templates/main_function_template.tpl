package main

import (
	"github.com/golangid/goruda/{{ .TargetDirectory }}/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCMD.Execute(); err != nil {
		logrus.Fatalf("Error initiate application: %v", err)
	}
}
