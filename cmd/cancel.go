package cmd

import (
	"github.com/spf13/cobra"
)

// Create the cancel command
var cmdCancel = &cobra.Command{
	Use:   "cancel WORKFLOW",
	Short: "Cancel a workflow",
	Long: `Cancel a Swif workflow.

Use "sw rm WORKFLOW" to delete a workflow.`,
	Example: `1. sw cancel my-workflow
2. sw cancel ana`,
	Run: runCancel,
}

func init() {
	cmdSW.AddCommand(cmdCancel)
}

func runCancel(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		exitNoWorkflow(cmd)
	}
	run("swif", "cancel", "-workflow", args[0])
}
