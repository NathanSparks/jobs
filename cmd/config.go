package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

type Config struct {
	JobPerDir       bool     `json:"jobPerDir"`
	InputDir        string   `json:"inputDir"`
	InputFilePrefix string   `json:"inputFilePrefix"`
	InputFileSuffix string   `json:"inputFileSuffix"`
	DirNoList       string   `json:"dirNoList"`
	DirNoDigits     string   `json:"dirNoDigits"`
	DirNoMin        int      `json:"dirNoMin"`
	DirNoMax        int      `json:"dirNoMax"`
	FileNoMin       int      `json:"fileNoMin"`
	FileNoMax       int      `json:"fileNoMax"`
	Workflow        string   `json:"workflow"`
	Phase           string   `json:"phase"`
	Project         string   `json:"project"`
	Track           string   `json:"track"`
	Name            string   `json:"name"`
	Tags            []string `json:"tags"`
	Cores           string   `json:"cores"`
	Disk            string   `json:"disk"`
	RAM             string   `json:"ram"`
	Time            string   `json:"time"`
	OS              string   `json:"os"`
	Inputs          []string `json:"inputs"`
	Outputs         []string `json:"outputs"`
	Stdout          string   `json:"stdout"`
	Stderr          string   `json:"stderr"`
	Shell           string   `json:"shell"`
	Command         string   `json:"command"`
}

func (c *Config) read(path string) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	json.Unmarshal(b, &c)

	if strings.HasSuffix(c.Shell, "/tcsh") {
		c.Shell += " -f"
	}
	wd := workDir()
	if strings.HasPrefix(c.Command, "script.") {
		c.Command = wd + "/" + c.Command
	}
	c.Command = c.Shell + " " + c.Command
}

func (c *Config) config(d, id string) {
	c.Name = strings.Replace(c.Name, "[jobID]", id, -1)
	var inputs, outputs []string
	for _, input := range c.Inputs {
		input = strings.Replace(input, "[dirNo]", d, -1)
		inputs = append(inputs, strings.Replace(input, "[jobID]", id, -1))
	}
	c.Inputs = inputs
	for _, output := range c.Outputs {
		if !c.JobPerDir {
			output = strings.Replace(output, "[dirNo]", d, -1)
		}
		outputs = append(outputs, strings.Replace(output, "[jobID]", id, -1))
	}
	c.Outputs = outputs
	c.Stdout = strings.Replace(c.Stdout, "[jobID]", id, -1)
	c.Stderr = strings.Replace(c.Stderr, "[jobID]", id, -1)
	input := c.InputFilePrefix + id + c.InputFileSuffix
	if c.JobPerDir {
		input = c.InputDir
		if strings.HasPrefix(input, "/mss/") {
			input = strings.Replace(input, "/mss/", "/cache/", 1)
		}
	}
	c.Command = strings.Replace(c.Command, "[input]", input, -1)
	c.Command = strings.Replace(c.Command, "[jobID]", id, -1)
}
