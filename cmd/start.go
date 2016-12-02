package cmd

import (
	"github.com/spf13/cobra"
)

// Create the start command
var cmdStart = &cobra.Command{
	Use:   "start WORKFLOW",
	Short: "Start/resume a workflow",
	Long:  `Start/resume a Swif workflow.`,
	Example: `1. sw start my-workflow
2. sw start ana`,
	Run: runStart,
}

var joblimit, phaselimit, errorlimit string

func init() {
	cmdSW.AddCommand(cmdStart)

	cmdStart.Flags().StringVarP(&joblimit, "joblimit", "j", "", "Job limit")
	cmdStart.Flags().StringVarP(&phaselimit, "phaselimit", "p", "", "Phase limit")
	cmdStart.Flags().StringVarP(&errorlimit, "errorlimit", "e", "", "Error limit")
}

func runStart(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		exitNoWorkflow(cmd)
	}
	var opts []string
	if joblimit != "" {
		opts = append(opts, "-joblimit", joblimit)
	}
	if phaselimit != "" {
		opts = append(opts, "-phaselimit", phaselimit)
	}
	if errorlimit != "" {
		opts = append(opts, "-errorlimit", errorlimit)
	}
	if len(opts) > 0 {
		c := []string{"run", "-workflow", args[0]}
		c = append(c, opts...)
		run("swif", c...)
	} else {
		run("swif", "run", "-workflow", args[0])
	}
}
