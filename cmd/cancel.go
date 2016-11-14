package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the cancel command
var cmdCancel = &cobra.Command{
	Use:   "cancel WORKFLOW",
	Short: "Cancel a workflow",
	Long: `Cancel a Swif workflow.

Use "sw rm WORKFLOW" to delete a workflow.

Usage example:
sw cancel ana
`,
	Run: runCancel,
}

func init() {
	cmdSW.AddCommand(cmdCancel)
}

func runCancel(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "Please specify a Swif workflow to cancel.\n")
		os.Exit(2)
	}
	run("swif", "cancel", "-workflow "+args[0])
}
