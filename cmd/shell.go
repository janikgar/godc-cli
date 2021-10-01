package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Run shell",
	Long:  "Run shell against inventory (or subset)",
	Run: func(cmd *cobra.Command, args []string) {
		getInventoryForShell()
	},
}

func openShell(host *Host, commandText string, user string, shellOutputChan chan ShellOutput) {
	key, err := ioutil.ReadFile("C:\\Users\\janik\\.ssh\\ansible_rsa")
	if err != nil {
		log.Fatalf("Error: could not open SSH private key: %s\n", err.Error())
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Error: after opening, could not parse SSH private key: %s\n", err.Error())
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshString := fmt.Sprintf("%v:%v", host.HostLocation, host.HostPort)

	client, err := ssh.Dial("tcp", sshString, config)
	if err != nil {
		log.Fatalf("Error: could not connect to %s as user %s; %s\n", sshString, user, err.Error())
	}
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Error: could not open SSH session: %s\n", err.Error())
	}
	output, err := session.CombinedOutput(commandText)
	defer session.Close()
	defer client.Close()
	shellOutputChan <- ShellOutput{
		Output: string(output),
		Host:   host,
		Error:  err,
	}
}

func getInventoryForShell() {
	var hostCount int
	shellOutputChan := make(chan ShellOutput)
	inventory := GetInventory(inventoryFile)
	for _, hostGroup := range inventory.Children {
		for _, host := range hostGroup.Hosts {
			host := host
			hostCount++
			go openShell(&host, commandString, user, shellOutputChan)
		}
	}
	for i := 0; i < hostCount; i++ {
		shellOutput := <-shellOutputChan
		fmt.Println(shellOutput.Display())
	}
}
