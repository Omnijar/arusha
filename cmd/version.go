package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/omnijar/arusha/config"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display this binary's version, build time and git hash of this build",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", config.Default.BuildVersion)
		fmt.Printf("Git Hash:   %s\n", config.Default.BuildHash)
		fmt.Printf("Build Time: %s\n", config.Default.BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
