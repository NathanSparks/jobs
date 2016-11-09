package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
		fmt.Fprint(os.Stderr, "Pass JSON config file as only argument.\n")
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

	if c.JobPerDir && strings.HasPrefix(c.InputDir, "/mss/") {
		fmt.Fprintln(os.Stderr, `One-job per directory mode is not supported for tape-library input.
The input files need to be precached or in a local directory.`)
		os.Exit(2)
	}

	type fileInfo struct {
		dirNo int
		path  string
	}

	var files []fileInfo
	Ndirs, Nfiles := 0, 0
	for dirNo := c.DirNoMin; dirNo <= c.DirNoMax; dirNo++ {
		dirNo_str := toString(dirNo, c.DirNoDigits)
		if c.DirNoList != "" && !in(inputList, dirNo_str) {
			continue
		}
		dir := c.InputDir
		if strings.Contains(dir, "[dirNo]") {
			dir = strings.Replace(dir, "[dirNo]", dirNo_str, -1)
		}
		if isPath(dir) {
			fileNo := 0
			for _, file := range readDir(dir) {
				if strings.HasPrefix(file, c.InputFilePrefix) &&
					strings.HasSuffix(file, c.InputFileSuffix) &&
					fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
					files = append(files, fileInfo{dirNo, dir + "/" + file})
					fileNo++
					if fileNo == 1 {
						Ndirs++
					}
					Nfiles++
				}
			}
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

	for _, file := range files {
		floc := filepath.Base(file.path)
		fid := strings.TrimPrefix(floc, c.InputFilePrefix)
		fid = strings.TrimSuffix(fid, c.InputFileSuffix)

		ds := toString(file.dirNo, c.DirNoDigits)

		if c.JobPerDir {
			fid = ds
		}
		c = c0.config(ds, fid)

		inputArg := ""
		if !c.JobPerDir {
			inputArg = "-input " + floc + " " + idp + file.path
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

	if !dryRun {
		fmt.Printf("%d jobs added.\n", Nfiles)
	}
	if Ndirs == 0 {
		fmt.Println("No jobs to submit.")
		os.Exit(0)
	}
	fmt.Printf("\n%d directories with input files were found.\nAverage of %v jobs/directory to submit.\n", Ndirs, float32(Nfiles)/float32(Ndirs))

	if submit {
		fmt.Printf("Submitting %s workflow ...\n", c.Workflow)
		run("swif", "run", "-workflow "+workflow)
	}
}

func toString(dirNo int, digits string) string {
	s := strconv.Itoa(dirNo)
	if digits != "" {
		s = fmt.Sprintf("%0"+digits+"s", s)
	}
	return s
}
