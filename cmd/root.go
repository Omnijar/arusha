package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/omnijar/arusha/config"
)

// RootCmd sets up the base configuration for the CLI without sub-commands.
var RootCmd = &cobra.Command{
	Use:   "arusha",
	Short: "Arusha is a cloud native high throughput auth service.",
}

// Execute adds all child commands to the root command sets flags appropriately.
func Execute() {
	if err := config.Initialize(); err != nil {
		log.Fatalln(err.Error())
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
