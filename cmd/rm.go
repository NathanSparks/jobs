package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the rm command
var cmdRemove = &cobra.Command{
	Use:   "rm",
	Short: "Remove a Swif workflow",
	Long: `Remove a Swif workflow.

Usage example:
sw rm -w ana
`,
	Run: runRemove,
}

func init() {
	cmdSW.AddCommand(cmdRemove)

	cmdRemove.Flags().StringVarP(&workflow, "workflow", "w", "", "Swif workflow")
}

func runRemove(cmd *cobra.Command, args []string) {
	if workflow == "" {
		fmt.Fprint(os.Stderr, "Please specify Swif workflow to act on.\n")
		os.Exit(2)
	}
	run("swif", "cancel", "-workflow "+workflow, "-delete")
}
