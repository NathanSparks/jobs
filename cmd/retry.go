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
	Long:  `Retry problem jobs of a Swif workflow.`,
	Example: `1. sw retry ana -p "SWIF-SYSTEM-ERROR AUGER-TIMEOUT"
2. sw retry my-workflow -p SWIF-SYSTEM-ERROR`,
	Run: runRetry,
}

var problems []string

func init() {
	cmdSW.AddCommand(cmdRetry)

	cmdRetry.Flags().StringSliceVarP(&problems, "problems", "p", nil, "Problem types (enclose multiple problems in quotes)")
}

func runRetry(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Required workflow argument is missing, or quotes are not enclosing multiple problems.\n")
		cmd.Usage()
		os.Exit(2)
	}
	if problems == nil {
		fmt.Fprintln(os.Stderr, "Required --problems flag is missing.\n")
		cmd.Usage()
		os.Exit(2)
	}
	swifArgs := []string{"retry-jobs", "-workflow", args[0], "-problems"}
	swifArgs = append(swifArgs, problems...)
	run("swif", swifArgs...)
}
