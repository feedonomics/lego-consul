package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/feedonomics/lego-consul/version"
)

var versionShort bool
var versionCmd = &cobra.Command{
	Use: `version`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionShort {
			fmt.Println(version.Get().String())
		} else {
			fmt.Println(version.Get().LongString())
		}
	},
}
