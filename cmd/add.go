package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Create the add command
var cmdAdd = &cobra.Command{
	Use:   "add [WORKFLOW]",
	Short: "Add jobs to or create a workflow",
	Long: `Add jobs to or create a Swif workflow.

The workflow is created if it does not already exist.

A JSON config file is used to configure the workflow.`,
	Example: `1. Create a new blank workflow called sim100.
    sw add sim100
2. Add jobs to a workflow by using a JSON configuration file.
    sw add -c config.json

Most of the config file fields correspond to "swif add-job" options.
Run "swif add-job -help" for details on those options.

If jobPerRun is false: Add one job per input file.
If jobPerRun is true : Add one job per run (multiple input files).

If jobPerRun is true and the input files are not from the tape library,
sw does not pass them to Auger to copy. The job script should copy the
input files to the local disk of the compute node for efficient I/O.
This choice allows the same simple script to be used for both types of
input: tape ("/mss/...") and non-tape ("/cache/...", etc.).

Generic [input] script argument
If jobPerRun is false: [input] = input filename
If jobPerRun is true : [input] = input directory

The job ID is determined from the allowed input filename forms: 
${PREFIX}${RUNNO}_${FILENO}${SUFFIX} or ${PREFIX}${RUNNO}${SUFFIX}
1. If jobPerRun is false: [jobID] = ${RUNNO}_${FILENO} or ${RUNNO}
2. If jobPerRun is true : [jobID] = ${RUNNO}

job tracks: debug, analysis, reconstruction, one_pass, simulation

More info:
  https://scicomp.jlab.org/docs/batch
  https://scicomp.jlab.org/docs/batch_job_faq
  https://scicomp.jlab.org/docs/swif`,
	Run: runAdd,
}

var dryRun, start bool
var config_file string

func init() {
	cmdSW.AddCommand(cmdAdd)

	cmdAdd.Flags().StringVarP(&config_file, "config", "c", "", "Path to JSON configuration file")
	cmdAdd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Print job commands and exit")
	cmdAdd.Flags().BoolVarP(&start, "start", "s", false, "Start the workflow after adding jobs")
}

func runAdd(cmd *cobra.Command, args []string) {
	if len(args) == 0 && config_file == "" {
		fmt.Fprintln(os.Stderr, "Required workflow argument and/or the -c option is missing.\n")
		cmd.Usage()
		os.Exit(2)
	}
	if len(args) > 0 && config_file == "" {
		run("swif", "create", "-workflow "+args[0])
		return
	}
	if config_file == "" {
		fmt.Fprintln(os.Stderr, "Use the -c option to specify a JSON config file.\n")
		cmd.Usage()
		os.Exit(2)
	}

	var c Config
	path := config_file
	if isPath(path) {
		c.read(path)
	} else {
		fmt.Fprintf(os.Stderr, "Path to config file does not exist:\n%s\n", path)
		os.Exit(2)
	}

	if len(args) > 0 {
		c.Workflow = args[0]
	}

	var runNoList []string
	if c.RunNoList != "" {
		if isPath(c.RunNoList) {
			runNoList = strings.Split(readFile(c.RunNoList), "\n")
		} else {
			fmt.Fprintf(os.Stderr, "Path to directory-number list does not exist:\n%s\n", c.RunNoList)
			os.Exit(2)
		}
	}

	if dryRun {
		fmt.Printf("Dry run: No jobs will be added to %s workflow.\n\n", c.Workflow)
	} else {
		fmt.Printf("Adding jobs to %s workflow ...\n", c.Workflow)
	}

	idp := "file:"
	if strings.HasPrefix(c.InputDir, "/mss/") {
		idp = "mss:"
	}
	c0 := c

	Ndirs, Nfiles, Nruns := 0, 0, 0

	switch {
	case c.InputDir == "":
		for runNo := c.RunNoMin; runNo <= c.RunNoMax; runNo++ {
			runNo_str := toString(runNo, c.RunNoDigits)
			if c.RunNoList != "" && !in(runNoList, runNo_str) {
				continue
			}
			Nruns++
			if c.JobPerRun {
				Nfiles++
				c.addJob(runNo_str, runNo_str, "")
			} else {
				for fileNo := c.FileNoMin; fileNo <= c.FileNoMax; fileNo++ {
					fileNo_str := toString(fileNo, c.FileNoDigits)
					file := runNo_str + "_" + fileNo_str
					Nfiles++
					c.addJob(runNo_str, file, "")
				}
			}
		}
		fmt.Printf("Nruns: %d\n", Nruns)
	default:
		var files []string
		sd := false
		if !strings.Contains(c.InputDir, "[runNo]") {
			sd = true
			Ndirs = 1
			files = readDir(c.InputDir)
		}
		switch c.JobPerRun {
		case false:
			for runNo := c.RunNoMin; runNo <= c.RunNoMax; runNo++ {
				runNo_str := toString(runNo, c.RunNoDigits)
				if c.RunNoList != "" && !in(runNoList, runNo_str) {
					continue
				}
				if !sd {
					c.InputDir = strings.Replace(c0.InputDir, "[runNo]", runNo_str, -1)
					files = readDir(c.InputDir)
				}
				fileNo := 0
				for _, file := range files {
					if strings.HasPrefix(file, c.InputFilePrefix) &&
						strings.HasSuffix(file, c.InputFileSuffix) &&
						fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
						if sd && !strings.Contains(file, runNo_str) {
							continue
						}
						c.addJob(runNo_str, file, idp)
						fileNo++
						if fileNo == 1 {
							Nruns++
						}
						Nfiles++
					}
				}
			}
		case true:
			for runNo := c.RunNoMin; runNo <= c.RunNoMax; runNo++ {
				runNo_str := toString(runNo, c.RunNoDigits)
				if c.RunNoList != "" && !in(runNoList, runNo_str) {
					continue
				}
				if !sd {
					c.InputDir = strings.Replace(c0.InputDir, "[runNo]", runNo_str, -1)
					files = readDir(c.InputDir)
				}
				fileNo := 0
				inputArgs, f0 := "", ""
				for _, file := range files {
					if strings.HasPrefix(file, c.InputFilePrefix) &&
						strings.HasSuffix(file, c.InputFileSuffix) &&
						fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
						if sd && !strings.Contains(file, runNo_str) {
							continue
						}
						fileNo++
						if fileNo == 1 {
							Nruns++
							if idp == "mss:" {
								inputArgs = "-input " + file + " " + idp + c.InputDir + "/" + file
							}
							f0 = file
						} else if idp == "mss:" {
							inputArgs = inputArgs + " -input " + file + " " + idp + c.InputDir + "/" + file
						}
						Nfiles++
					}
				}
				if fileNo == 0 {
					continue
				}
				c.addJob(runNo_str, f0, inputArgs)
			}
		}
		if !sd {
			Ndirs = Nruns
		}
		fmt.Printf("%d input files were found for %d runs in %d directories.\n", Nfiles, Nruns, Ndirs)
	}

	if Nfiles == 0 {
		fmt.Println("No jobs to add.")
		os.Exit(0)
	}
	Njobs := Nfiles
	if c.JobPerRun {
		Njobs = Nruns
	}
	fmt.Printf("%d jobs to add (Njobs/Nruns = %v).\n", Njobs, float32(Njobs)/float32(Nruns))

	if !dryRun && start {
		fmt.Printf("Starting %s workflow ...\n", c.Workflow)
		run("swif", "run", "-workflow", c.Workflow)
	}
}

func (c Config) addJob(runNo, file, idp string) {
	fid := strings.TrimPrefix(file, c.InputFilePrefix)
	fid = strings.TrimSuffix(fid, c.InputFileSuffix)

	if c.JobPerRun {
		fid = runNo
	}
	c.config(runNo, fid)

	inputArgs := ""
	if !c.JobPerRun && c.InputDir != "" {
		inputArgs = "-input " + file + " " + idp + c.InputDir + "/" + file
	} else {
		inputArgs = idp
	}
	for _, input := range c.Inputs {
		if inputArgs == "" {
			inputArgs = "-input " + input
		} else {
			inputArgs = inputArgs + " -input " + input
		}
	}
	outputArgs := ""
	for _, output := range c.Outputs {
		if outputArgs == "" {
			outputArgs = "-output " + output
		} else {
			outputArgs = outputArgs + " -output " + output
		}
	}
	tagArgs := ""
	for _, tag := range c.Tags {
		if tagArgs == "" {
			tagArgs = "-tag " + tag
		} else {
			tagArgs = tagArgs + " -tag " + tag
		}
	}

	args := []string{"add-job", "-create", "-workflow", c.Workflow, "-project", c.Project, "-track", c.Track}
	var opts []string
	flags := []flg{{"-name", c.Name}, {"-cores", c.Cores}, {"-disk", c.Disk}, {"-ram", c.RAM}, {"-phase", c.Phase},
		{"-time", c.Time}, {"-os", c.OS}, {"-stdout", c.Stdout}, {"-stderr", c.Stderr}}
	for _, v := range flags {
		opts = maybeAppend(opts, v)
	}
	if tagArgs != "" {
		opts = append(opts, strings.Split(tagArgs, " ")...)
	}
	if inputArgs != "" {
		opts = append(opts, strings.Split(inputArgs, " ")...)
	}
	if outputArgs != "" {
		opts = append(opts, strings.Split(outputArgs, " ")...)
	}
	opts = append(opts, strings.Split(c.Command, " ")...)
	args = append(args, opts...)

	if dryRun {
		fmt.Println("swif " + strings.Join(args, " ") + "\n")
	} else {
		run("swif", args...)
	}
}

func toString(n int, digits string) string {
	s := strconv.Itoa(n)
	if digits != "" {
		s = fmt.Sprintf("%0"+digits+"s", s)
	}
	return s
}

func maybeAppend(a []string, b flg) []string {
	var c []string
	c = append(c, a...)
	if b.val != "" {
		c = append(c, b.name, b.val)
	}
	return c
}

type flg struct {
	name string
	val  string
}
