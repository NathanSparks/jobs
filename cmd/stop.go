package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Create the stop command
var cmdStop = &cobra.Command{
	Use:   "stop WORKFLOW",
	Short: "Stop (pause) a workflow",
	Long: `Stop (pause) a Swif workflow.

Usage example:
sw stop ana
`,
	Run: runStop,
}

func init() {
	cmdSW.AddCommand(cmdStop)
}

func runStop(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, `Required "workflow" argument is missing.
Run "sw help stop" for usage details.`)
		os.Exit(2)
	}
	run("swif", "pause", "-workflow", args[0])
}
