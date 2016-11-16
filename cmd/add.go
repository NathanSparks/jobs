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

A JSON config file is used to configure the workflow.

job tracks: debug, analysis, reconstruction, one_pass, simulation

Usage examples:
1. Create a new blank workflow called sim100.
    sw add sim100
2. Add jobs to a workflow by using a JSON configuration file.
    sw add -c config.json
`,
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
		fmt.Fprintln(os.Stderr, `"sw add" requires a workflow as an argument and/or the "-c" option.
Run "sw help add" for usage details.`)
		os.Exit(2)
	}
	if len(args) > 0 && config_file == "" {
		run("swif", "create", "-workflow "+args[0])
		return
	}
	if config_file == "" {
		fmt.Fprintln(os.Stderr, `Please use the "-c" option to specify the JSON config file.
Run "sw help add" for usage details.`)
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

	var dirNoList []string
	if c.DirNoList != "" {
		if isPath(c.DirNoList) {
			dirNoList = strings.Split(readFile(c.DirNoList), "\n")
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
			if c.DirNoList != "" && !in(dirNoList, dirNo_str) {
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
			if c.DirNoList != "" && !in(dirNoList, dirNo_str) {
				continue
			}
			c.InputDir = strings.Replace(c0.InputDir, "[dirNo]", dirNo_str, -1)
			fileNo := 0
			inputArgs, f0 := "", ""
			for _, file := range readDir(c.InputDir) {
				if strings.HasPrefix(file, c.InputFilePrefix) &&
					strings.HasSuffix(file, c.InputFileSuffix) &&
					fileNo >= c.FileNoMin && fileNo <= c.FileNoMax {
					fileNo++
					if fileNo == 1 {
						Ndirs++
						inputArgs = "-input " + file + " " + idp + c.InputDir + "/" + file
						f0 = file
					} else {
						inputArgs = inputArgs + " -input " + file + " " + idp + c.InputDir + "/" + file
					}
					Nfiles++
				}
			}
			if fileNo == 0 {
				continue
			}
			c.addJob(dirNo, f0, inputArgs)
		}
	}

	fmt.Printf("%d input files were found in %d directories.\n", Nfiles, Ndirs)
	if Ndirs == 0 {
		fmt.Println("No jobs to add.")
		os.Exit(0)
	}
	Njobs := Nfiles
	if c.JobPerDir {
		Njobs = Ndirs
	}
	fmt.Printf("%d jobs to add (Njobs/Ndirs = %v).\n", Njobs, float32(Njobs)/float32(Ndirs))

	if !dryRun && start {
		fmt.Printf("Starting %s workflow ...\n", c.Workflow)
		run("swif", "run", "-workflow", c.Workflow)
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

	inputArgs := ""
	if !c.JobPerDir {
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

func toString(dirNo int, digits string) string {
	s := strconv.Itoa(dirNo)
	if digits != "" {
		s = fmt.Sprintf("%0"+digits+"s", s)
	}
	return s
}

func maybeAppend(a []string, b flg) []string {
	var c []string
	c = append(c, a...)
	if b.val != "" {
		c = append(c, []string{b.name, b.val}...)
	}
	return c
}

type flg struct {
	name string
	val  string
}
