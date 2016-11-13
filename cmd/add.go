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
	Use:   "add CONFIG_FILE",
	Short: "Add jobs to a Swif workflow",
	Long: `Add jobs to a Swif workflow.

The workflow is created if it does not already exist.

Pass a JSON config file as the only argument.

job tracks: debug, analysis, reconstruction, one_pass, simulation

Usage example:
sw add config.json
`,
	Run: runAdd,
}

var dryRun, submit bool

func init() {
	cmdSW.AddCommand(cmdAdd)

	cmdAdd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Print job commands and exit")
	cmdAdd.Flags().BoolVarP(&submit, "submit", "s", false, "Submit the Swif workflow after adding jobs")
}

func runAdd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "Pass a JSON config file as the only argument.\n")
		os.Exit(2)
	}

	wd := workDir()

	var c Config
	path := wd + "/" + args[0]
	if isPath(path) {
		c.read(path)
	} else {
		fmt.Fprintf(os.Stderr, "Path to config file does not exist:\n%s\n", path)
		os.Exit(2)
	}

	var inputList []string
	if c.DirNoList != "" {
		if isPath(c.DirNoList) {
			inputList = strings.Split(readFile(c.DirNoList), "\n")
		} else {
			fmt.Fprintf(os.Stderr, "Path to directory-number list does not exist:\n%s\n", c.DirNoList)
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

	Ndirs, Nfiles := 0, 0
	switch c.JobPerDir {
	case false:
		for dirNo := c.DirNoMin; dirNo <= c.DirNoMax; dirNo++ {
			dirNo_str := toString(dirNo, c.DirNoDigits)
			if c.DirNoList != "" && !in(inputList, dirNo_str) {
				continue
			}
			c.InputDir = strings.Replace(c0.InputDir, "[dirNo]", dirNo_str, -1)
			fileNo := 0
			for _, file := range readDir(c.InputDir) {
				if strings.HasPrefix(file, c.InputFilePrefix) &&
					strings.HasSuffix(file, c.InputFileSuffix) &&
					fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
					c.addJob(dirNo, file, idp)
					fileNo++
					if fileNo == 1 {
						Ndirs++
					}
					Nfiles++
				}
			}
		}
	case true:
		for dirNo := c.DirNoMin; dirNo <= c.DirNoMax; dirNo++ {
			dirNo_str := toString(dirNo, c.DirNoDigits)
			if c.DirNoList != "" && !in(inputList, dirNo_str) {
				continue
			}
			c.InputDir = strings.Replace(c0.InputDir, "[dirNo]", dirNo_str, -1)
			fileNo := 0
			inputArg, f0 := "", ""
			for _, file := range readDir(c.InputDir) {
				if strings.HasPrefix(file, c.InputFilePrefix) &&
					strings.HasSuffix(file, c.InputFileSuffix) &&
					fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
					fileNo++
					if fileNo == 1 {
						Ndirs++
						inputArg = "-input " + file + " " + idp + c.InputDir + "/" + file
						f0 = file
					} else {
						inputArg = inputArg + " -input " + file + " " + idp + c.InputDir + "/" + file
					}
					Nfiles++
				}
			}
			if fileNo == 0 {
				continue
			}
			c.addJob(dirNo, f0, inputArg)
		}
	}

	if !dryRun {
		fmt.Printf("%d input files were found.\n", Nfiles)
	}
	if Ndirs == 0 {
		fmt.Println("No jobs to submit.")
		os.Exit(0)
	}
	Njobs := Nfiles
	if c.JobPerDir {
		Njobs = Ndirs
	}
	fmt.Printf("\n%d directories with input files were found.\nAverage of %v jobs/directory to submit.\n", Ndirs, float32(Njobs)/float32(Ndirs))

	if submit {
		fmt.Printf("Submitting %s workflow ...\n", c.Workflow)
		run("swif", "run", "-workflow "+workflow)
	}
}

func (c Config) addJob(dirNo int, file, idp string) {
	fid := strings.TrimPrefix(file, c.InputFilePrefix)
	fid = strings.TrimSuffix(fid, c.InputFileSuffix)

	ds := toString(dirNo, c.DirNoDigits)

	if c.JobPerDir {
		fid = ds
	}
	c.config(ds, fid)

	inputArg := ""
	if !c.JobPerDir {
		inputArg = "-input " + file + " " + idp + c.InputDir + "/" + file
	} else {
		inputArg = idp
	}
	for _, input := range c.Inputs {
		if inputArg == "" {
			inputArg = "-input " + input
		} else {
			inputArg = inputArg + " -input " + input
		}
	}
	outputArg := ""
	for _, output := range c.Outputs {
		if outputArg == "" {
			outputArg = "-output " + output
		} else {
			outputArg = outputArg + " -output " + output
		}
	}

	args := "add-job -create -workflow " + c.Workflow + " -project " + c.Project +
		" -track " + c.Track + " -cores " + c.Cores + " -disk " + c.Disk + " -ram " + c.RAM +
		" -time " + c.Time + " -os " + c.OS + " " + inputArg + " " + outputArg + " -stdout " +
		c.Stdout + " -stderr " + c.Stderr + " -name " + c.Name + " " + c.Command

	if dryRun {
		fmt.Println("swif " + args + "\n")
	} else {
		run("swif", args)
	}
}

func toString(dirNo int, digits string) string {
	s := strconv.Itoa(dirNo)
	if digits != "" {
		s = fmt.Sprintf("%0"+digits+"s", s)
	}
	return s
}
