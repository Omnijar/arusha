package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/omnijar/arusha/server"
)

// hostCmd represents the host command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Start the HTTP/2 host service",
	Long:  `Starts all HTTP/2 APIs and connects to a database backend.`,
	Run:   server.RunHost(),
}

func init() {
	RootCmd.AddCommand(hostCmd)
}
