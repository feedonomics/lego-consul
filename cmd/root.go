package cmd

import (
	"time"

	"github.com/go-acme/lego/v4/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/ottodashadow/lego-consul/config"
)

var RootCmd = &cobra.Command{
	Use: "lego-consul",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Logger = logrus.StandardLogger()
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})

		if config.Quiet {
			logrus.SetLevel(logrus.WarnLevel)
		}
	},
}
