package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the retry command
var cmdRetry = &cobra.Command{
	Use:   "retry",
	Short: "Retry problem jobs",
	Long: `Retry problem jobs.

Usage example:
sw retry -w ana -p SWIF-SYSTEM-ERROR
`,
	Run: runRetry,
}

var workflow string
var problems string

func init() {
	cmdSW.AddCommand(cmdRetry)

	cmdRetry.Flags().StringVarP(&workflow, "workflow", "w", "", "Swif workflow")
	cmdRetry.Flags().StringVarP(&problems, "problems", "p", "", "Problem types")
}

func runRetry(cmd *cobra.Command, args []string) {
	if workflow == "" {
		fmt.Fprint(os.Stderr, "Please specify Swif workflow to act on.\n")
		os.Exit(2)
	}
	if problems == "" {
		fmt.Fprint(os.Stderr, "Please specify problem types.\n")
		os.Exit(2)
	}
	run("swif", "retry-jobs", "-workflow "+workflow, "-problems "+problems)
}
