package cmd

import (
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the server",
	Long:  `Stop the server`,
}

func stop() {

}
