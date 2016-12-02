package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the rm command
var cmdRemove = &cobra.Command{
	Use:   "rm WORKFLOW",
	Short: "Remove a workflow",
	Long: `Remove a Swif workflow.

Use "sw cancel WORKFLOW" to cancel a workflow without deleting it.`,
	Example: `1. sw rm my-workflow
2. sw rm ana`,
	Run: runRemove,
}

func init() {
	cmdSW.AddCommand(cmdRemove)
}

func runRemove(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		exitNoWorkflow(cmd)
	}
	run("swif", "cancel", "-workflow", args[0], "-delete")
}

func exitNoWorkflow(cmd *cobra.Command) {
	fmt.Fprintln(os.Stderr, "Required workflow argument is missing.\n")
	cmd.Usage()
	os.Exit(2)
}
