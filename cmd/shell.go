package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

func findSshPrivateKeys() []ssh.Signer {
	var possibleKeys []ssh.Signer
	sshDirName := keyPath
	if sshDirName == "" {
		log.Println("No key directory given; searching default SSH key location")
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error: could not use home dir %s: %s\n", userHomeDir, err.Error())
		}
		userHomeDirRead, err := os.ReadDir(userHomeDir)
		if err != nil {
			log.Fatalf("Error: could not read home dir %s: %s\n", userHomeDir, err.Error())
		}
		var sshDirEntry os.DirEntry
		for _, dirItem := range userHomeDirRead {
			if dirItem.IsDir() && dirItem.Name() == ".ssh" {
				sshDirEntry = dirItem
			}
		}
		if sshDirEntry == nil {
			log.Fatalf("Error: no user SSH directory found under %s\n", userHomeDir)
		}
		sshDirName = filepath.Join(userHomeDir, ".ssh")
	}
	sshDir, err := os.ReadDir(sshDirName)
	if err != nil {
		log.Fatalf("Error: could not open SSH dir %s: %s\n", sshDirName, err.Error())
	}
	for _, sshDirItem := range sshDir {
		if !sshDirItem.IsDir() {
			possibleKeyPath := filepath.Join(sshDirName, sshDirItem.Name())
			possibleKey, err := ioutil.ReadFile(possibleKeyPath)
			if err != nil {
				log.Fatalf("Error: could not open file %s: %s\n", possibleKeyPath, err.Error())
			}
			key, _ := ssh.ParsePrivateKey(possibleKey)
			possibleKeys = append(possibleKeys, key)
		}
	}
	return possibleKeys
}

func openShell(host *Host, commandText string, user string, shellOutputChan chan ShellOutput) {
	var (
		session *ssh.Session
		client  *ssh.Client
		err     error
	)
	possibleKeys := findSshPrivateKeys()

	for _, key := range possibleKeys {
		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(key),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		sshString := fmt.Sprintf("%v:%v", host.HostLocation, host.HostPort)

		defer func(shellOutputChan chan ShellOutput) {
			if r := recover(); r != nil {
				shellOutputChan <- ShellOutput{
					Output: "",
					Host:   host,
					Error:  fmt.Errorf("could not connect with user: %s", config.User).Error(),
				}
			}
		}(shellOutputChan)
		client, err = ssh.Dial("tcp", sshString, config)
		if err != nil {
			client = nil
			continue
		}
		break
	}
	if client == nil {
		log.Fatalln("Error: could not find a usable SSH key")
	}
	session, err = client.NewSession()
	if err != nil {
		log.Fatalf("Error: could not open SSH session: %s\n", err.Error())
	}
	output, err := session.CombinedOutput(commandText)
	defer session.Close()
	defer client.Close()
	shellOutputChan <- ShellOutput{
		Output: string(output),
		Host:   host,
		Error:  err.Error(),
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
