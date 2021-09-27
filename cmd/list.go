package cmd

import (
	"fmt"

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

func getHosts(inputYaml map[interface{}]interface{}, vars map[interface{}]interface{}) []Host {
	var hosts []Host
	localVars := make(map[interface{}]interface{})
	for varKey, varVal := range vars {
		localVars[varKey] = varVal
	}
	if v, ok := inputYaml["vars"]; ok {
		if v, ok := v.(map[interface{}]interface{}); ok {
			for varK, varV := range v {
				localVars[varK] = varV
			}
		}
	}
	if v, ok := inputYaml["hosts"]; ok {
		if v, ok := v.(map[interface{}]interface{}); ok {
			for hostK, hostV := range v {
				if hostK, ok := hostK.(string); ok {
					hostLocation := hostK
					hostPort := 22
					if hostV, ok := hostV.(map[interface{}]interface{}); ok {
						if altHost, ok := hostV["ansible_host"]; ok {
							if altHost, ok := altHost.(string); ok {
								hostLocation = altHost
							}
						}
						if altPort, ok := hostV["ansible_port"]; ok {
							if altPort, ok := altPort.(int); ok {
								hostPort = altPort
							}
						}
					}
					host := Host{
						HostName:     hostK,
						HostLocation: hostLocation,
						HostPort:     hostPort,
						Vars:         localVars,
					}
					hosts = append(hosts, host)
				}
			}
		}
	}
	return hosts
}

func getHostGroups(inputYaml map[interface{}]interface{}, vars map[interface{}]interface{}) []HostGroup {
	var hostGroups []HostGroup
	for k, v := range inputYaml {
		var hosts []Host
		if k, ok := k.(string); ok {
			if v, ok := v.(map[interface{}]interface{}); ok {
				if h, ok := v["hosts"]; ok {
					if _, ok = h.(map[interface{}]interface{}); ok {
						hosts = getHosts(v, vars)
					}
				}
			}
			hostGroup := HostGroup{
				GroupName: k,
				Hosts:     hosts,
				Vars:      vars,
			}
			hostGroups = append(hostGroups, hostGroup)
		}
	}
	return hostGroups
}

func listInventory() {
	parseInventoryYml(inventoryFile)
	var allHosts AllHosts
	var hostGroups []HostGroup
	var vars map[interface{}]interface{}
	allYaml, ok := inventoryYaml["all"]
	if ok {
		if allYaml, ok := allYaml.(map[interface{}]interface{}); ok {
			if yamlVars, ok := allYaml["vars"]; ok {
				if yamlVars, ok := yamlVars.(map[interface{}]interface{}); ok {
					allHosts.GlobalVars = yamlVars
					vars = yamlVars
				}
			}
			if children, ok := allYaml["children"]; ok {
				if children, ok := children.(map[interface{}]interface{}); ok {
					hostGroups = getHostGroups(children, vars)
					allHosts.Children = hostGroups
				}
			}
		}
	}
	fmt.Println(allHosts.Display())
}
