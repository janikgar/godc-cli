package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func display(a interface{}) string {
	var output string
	switch outputFmt {
	case "yaml":
		yamlForm, err := yaml.Marshal(a)
		if err != nil {
			log.Fatalf("Error: Could not marshal YAML: %s\n", err.Error())
		}
		output = string(yamlForm)
	case "json":
		jsonForm, err := jsoniter.MarshalIndent(a, "", "  ")
		if err != nil {
			log.Fatalf("Error: Could not marshal JSON: %s\n", err.Error())
		}
		output = string(jsonForm)
	case "struct":
		output = fmt.Sprintf("%+v\n", a)
	}
	return output
}

type Host struct {
	HostName     string
	HostLocation string `json:"ansible_host" yaml:"ansible_host"`
	HostPort     int    `json:"ansible_port" yaml:"ansible_port"`
	Groups       []string
	Vars         map[interface{}]interface{}
}

func (h Host) Display() string {
	var allVars string
	var i int
	for k, v := range h.Vars {
		if i > 0 {
			allVars = allVars + ","
		}
		allVars = allVars + fmt.Sprintf(" %v: %v", k, v)
		i++
	}
	hostObj := fmt.Sprintf("{ HostName: %s, HostLocation: %s, HostPort: %d, Groups: [ %s ], Vars: {%s} }", h.HostName, h.HostLocation, h.HostPort, strings.Join(h.Groups, ", "), allVars)
	return hostObj
}

type HostGroup struct {
	GroupName string
	Hosts     []Host                      `json:"hosts"`
	Vars      map[interface{}]interface{} `json:"vars"`
}

func (hg HostGroup) Display() string {
	return fmt.Sprintf("{ GroupName: %+v, Hosts: %+v, Vars: %+v", hg.GroupName, hg.Hosts, hg.Vars)
}

type AllHosts struct {
	Children   []HostGroup                 `json:"children"`
	GlobalVars map[interface{}]interface{} `json:"vars" yaml:"vars"`
}

func (a AllHosts) Display() string {
	return display(a)
}

type TopLevel struct {
	All AllHosts `json:"all"`
}

type ShellOutput struct {
	Host   *Host
	Output string
	Error  error
}

func (a ShellOutput) Display() string {
	if outputFmt == "text" {
		var outputLines []string
		var maxLength int
		var spacerLine []string
		inputLines := strings.Split(a.Output, "\n")
		if a.Error != nil {
			inputLines = []string{a.Error.Error()}
		}
		for _, inputLine := range inputLines {
			hostname := a.Host.HostName
			outputLine := fmt.Sprintf("| %s | %s", hostname, inputLine)
			if len(outputLine) > maxLength {
				maxLength = len(outputLine)
			}
			outputLines = append(outputLines, outputLine)
		}
		for i := 0; i < maxLength+1; i++ {
			if i == 0 || i == maxLength {
				spacerLine = append(spacerLine, "+")
				continue
			}
			spacerLine = append(spacerLine, "-")
		}
		spacerLineText := strings.Join(spacerLine, "")
		outputLines = append(outputLines, spacerLineText)
		return strings.Join(outputLines, "\n")
	}
	return display(a)
}

var (
	rootCmd = &cobra.Command{
		Use:   "godc",
		Short: "godc is a multitool for connecting to your data center machines",
		Run: func(cmd *cobra.Command, args []string) {
			parseInventoryYml(inventoryFile)
		},
	}
	inventoryFile string
	outputFmt     string
)

var (
	inventoryYaml = make(map[string]interface{})
	commandString string
	user          string
	keyPath       string
)

func parseInventoryYml(inventoryFile string) {
	var inventoryContent []byte
	inventoryContent, err := ioutil.ReadFile(inventoryFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %s\n", inventoryFile, err.Error())
	}
	err = yaml.Unmarshal(inventoryContent, &inventoryYaml)
	if err != nil {
		log.Fatalf("Error unmarshalling file %s: %s\n", inventoryFile, err.Error())
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inventory", "i", "inventory.yml", "Ansible Inventory file")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "yaml", "Output Format")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringVarP(&commandString, "command", "c", "uname -a", "Shell command to send")
	shellCmd.Flags().StringVarP(&user, "user", "u", "nobody", "Shell user")
	shellCmd.Flags().StringVarP(&keyPath, "key", "k", "", "SSH key directory")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
