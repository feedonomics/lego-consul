package cmd

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/feedonomics/lego-consul/config"
	"github.com/feedonomics/lego-consul/paths"
	"github.com/feedonomics/lego-consul/types"
	"github.com/feedonomics/lego-consul/utility"
)

var manageSANsCmd = &cobra.Command{
	Use: `sans`,
}

var addSANsCmd = &cobra.Command{
	Use:   `add`,
	Short: `Add subject alternate names to an existing domain renewal configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Domain == `` {
			return errors.New(`certificate domain is required`)
		}
		CleanSANs := utility.ParseSANs(SANs)
		if len(CleanSANs) == 0 {
			return errors.New(`at lease on subject alternate name is required`)
		}

		RenewalCfgPath := fmt.Sprintf(`%s/renewal/%s.json`, config.Path, Domain)
		var RenewalCfg types.Renewal
		if err := paths.ReadFileJSON(RenewalCfgPath, &RenewalCfg); err != nil {
			return err
		}

		var Changed bool
		for _, CleanSAN := range CleanSANs {
			if !utility.ExistsStrings(RenewalCfg.SANs, CleanSAN) &&
				CleanSAN != Domain {
				RenewalCfg.SANs = append(RenewalCfg.SANs, CleanSAN)
				Changed = true
			}
		}

		if Changed {
			logrus.Infof(`[INFO] [%s] renewal: writing renewal configuration`, Domain)
			if err := paths.WriteFileJSON(RenewalCfgPath, RenewalCfg, 0644); err != nil {
				return err
			}
		}

		return nil
	},
}
