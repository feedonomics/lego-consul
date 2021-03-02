package main

import (
	"github.com/sirupsen/logrus"

	"github.com/ottodashadow/lego-consul/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
