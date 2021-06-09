package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"time"

	"github.com/go-acme/lego/v4/log"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/feedonomics/lego-consul/config"
)

type consulCfg struct {
	Token string `json:"token"`
}

var RootCmd = &cobra.Command{
	Use: "lego-consul",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.Logger = logrus.StandardLogger()
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})

		if config.Quiet {
			logrus.SetLevel(logrus.WarnLevel)
		}

		if config.Syslog {
			Log, err := syslog.New(syslog.LOG_NOTICE, `lego_acme`)
			if err != nil {
				return err
			}
			logrus.SetOutput(Log)
		}

		ConsulCfgPath := fmt.Sprintf(`%s/conf.d/consul.json`, config.Path)
		if _, err := os.Stat(ConsulCfgPath); err == nil {
			Data, err := ioutil.ReadFile(ConsulCfgPath)
			if err != nil {
				return err
			}
			var cfg consulCfg
			if err := json.Unmarshal(Data, &cfg); err != nil {
				return err
			}

			if os.Getenv(api.HTTPTokenEnvName) == `` && cfg.Token != `` {
				if err := os.Setenv(api.HTTPTokenEnvName, cfg.Token); err != nil {
					return err
				}
			}
		}

		return nil
	},
}
