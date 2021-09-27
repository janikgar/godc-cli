package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Host struct {
	HostName     string
	HostLocation string `json:"ansible_host"`
	HostPort     int    `json:"ansible_port"`
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
	output := fmt.Sprintf("%+v\n", a)
	// jsonForm, err := jsoniter.MarshalIndent(a, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error: Could not marshal JSON: %s\n", err.Error())
	// }
	// return string(jsonForm)
	// yamlForm, err := yaml.Marshal(a)
	// if err != nil {
	// 	log.Fatalf("Error: Could not marshal YAML: %s\n", err.Error())
	// }
	return string(output)
}

type TopLevel struct {
	All AllHosts `json:"all"`
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
)

var inventoryYaml = make(map[string]interface{})

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
	rootCmd.AddCommand(listCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
