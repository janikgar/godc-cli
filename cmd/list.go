package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List inventory",
	Long:  "List inventory from imported YAML",
	Run: func(cmd *cobra.Command, args []string) {
		listInventory()
	},
}

func appendAndRecurse(k interface{}, v interface{}, groups []string) (map[interface{}]interface{}, []string) {
	var newV map[interface{}]interface{}
	if v != nil {
		newV = v.(map[interface{}]interface{})
	}
	newK := k.(string)
	groups = append(groups, newK)
	return newV, groups
}

func findHosts(input map[interface{}]interface{}, groups []string, vars []map[interface{}]interface{}) {
	for k, v := range input {
		if k == "vars" {
			if v != nil {
				v := v.(map[interface{}]interface{})
				vars = append(vars, v)
				input["vars"] = nil
			}
		}
	}
	for k, v := range input {
		switch k {
		case "vars":
			var newV map[interface{}]interface{}
			if v != nil {
				newV = v.(map[interface{}]interface{})
				vars = append(vars, newV)
			}
			findHosts(input, groups, vars)
		case "hosts":
			if v != nil {
				fmt.Printf("groups: [%s]\n", strings.Join(groups, ", "))
				fmt.Printf("vars: [%s]\n", vars)
				fmt.Println(k, ": ", v)
				fmt.Println("-----")
			}
		default:
			newV, groups := appendAndRecurse(k, v, groups)
			findHosts(newV, groups, vars)
		}
		input[k] = nil
		switch v := v.(type) {
		case nil:
			break
		case map[interface{}]interface{}:
			if len(v) == 0 {
				break
			}
		}
		if v == nil {
			break
		}
	}
}

func listInventory() {
	parseInventoryYml(inventoryFile)
	// fmt.Println(inventoryYaml)
	findHosts(inventoryYaml, nil, nil)
	// fmt.Println(reflect.TypeOf(inventoryYaml))
	// jsonOutBytes, err := json.Marshal(inventoryYaml)
	// jsonOutBytes, err := json.MarshalIndent(inventoryYaml, "", "  ")
	// if err != nil {
	// 	log.Fatalf("Error marshalling JSON: %s\n", err.Error())
	// }
	// fmt.Println(string(jsonOutBytes))
}
