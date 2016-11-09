package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Create the sw command
var cmdSW = &cobra.Command{
	Use:   "sw [COMMAND] [ARGS]",
	Short: "A tool for managing Swif workflows",
	Long: `
sw is a tool for managing Swif workflows.
`,
}

// Execute a sw command
func Execute() {
	if err := cmdSW.Execute(); err != nil {
		os.Exit(-1)
	}
}
