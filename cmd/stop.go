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

var now bool

func init() {
	cmdSW.AddCommand(cmdStop)

	cmdStop.Flags().BoolVarP(&now, "now", "n", false, "Cancel/recall running jobs")
}

func runStop(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, `Required workflow argument is missing.

Run "sw stop -h" for usage details.`)
		os.Exit(2)
	}
	if now {
		run("swif", "pause", "-workflow", args[0], "-now")
	} else {
		run("swif", "pause", "-workflow", args[0])
	}
}
