package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the start command
var cmdStart = &cobra.Command{
	Use:   "start WORKFLOW",
	Short: "Start/resume a workflow",
	Long: `Start/resume a Swif workflow.

Usage example:
sw start ana
`,
	Run: runStart,
}

func init() {
	cmdSW.AddCommand(cmdStart)
}

func runStart(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "Please specify a Swif workflow to start.\n")
		os.Exit(2)
	}
	run("swif", "run", "-workflow "+args[0])
}
