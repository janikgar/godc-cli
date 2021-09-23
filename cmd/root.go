package cmd

import (
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Host struct {
	HostName     string
	HostLocation string `json:"ansible_host"`
	HostPort     int    `json:"ansible_port"`
}

type HostGroup struct {
	GroupName string
	Hosts     map[string]Host        `json:"hosts"`
	Vars      map[string]interface{} `json:"vars"`
}

type AllHosts struct {
	Children   map[string]HostGroup   `json:"children"`
	GlobalVars map[string]interface{} `json:"vars"`
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

// var inventoryYaml = make(map[string]AllHosts)
var inventoryYaml = make(map[interface{}]interface{})

func parseInventoryYml(inventoryFile string) {
	var inventoryContent []byte
	inventoryContent, err := ioutil.ReadFile(inventoryFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %s\n", inventoryFile, err.Error())
	}
	err = yaml.Unmarshal(inventoryContent, inventoryYaml)
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
