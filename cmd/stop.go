package cmd

import (
	"github.com/spf13/cobra"
)

// Create the stop command
var cmdStop = &cobra.Command{
	Use:   "stop WORKFLOW",
	Short: "Stop (pause) a workflow",
	Long:  `Stop (pause) a Swif workflow.`,
	Example: `1. sw stop my-workflow
2. sw stop ana`,
	Run: runStop,
}

var now bool

func init() {
	cmdSW.AddCommand(cmdStop)

	cmdStop.Flags().BoolVarP(&now, "now", "n", false, "Cancel/recall running jobs")
}

func runStop(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		exitNoWorkflow(cmd)
	}
	if now {
		run("swif", "pause", "-workflow", args[0], "-now")
	} else {
		run("swif", "pause", "-workflow", args[0])
	}
}
