package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Create the version command
var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Show the sw version number",
	Long: `
Show the sw version number.
`,
	Run: runVersion,
}

const VERSION = "dev"

func init() {
	cmdSW.AddCommand(cmdVersion)
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("sw version %s\n", VERSION)
}
