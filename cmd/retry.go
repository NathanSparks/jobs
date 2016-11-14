package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the retry command
var cmdRetry = &cobra.Command{
	Use:   "retry WORKFLOW",
	Short: "Retry problem jobs of a workflow",
	Long: `Retry problem jobs of a Swif workflow.

Usage example:
sw retry ana -p SWIF-SYSTEM-ERROR
`,
	Run: runRetry,
}

var problems string

func init() {
	cmdSW.AddCommand(cmdRetry)

	cmdRetry.Flags().StringVarP(&problems, "problems", "p", "", "Problem types (enclose multiple problems in quotes)")
}

func runRetry(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "Please specify a Swif workflow to act on.\n")
		os.Exit(2)
	}
	if problems == "" {
		fmt.Fprint(os.Stderr, "Please specify problem types.\n")
		os.Exit(2)
	}
	run("swif", "retry-jobs", "-workflow "+args[0], "-problems "+problems)
}
