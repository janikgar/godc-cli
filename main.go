package main

import "github.com/janikgar/godc-cli/cmd"

func main() {
	cmd.Execute()
	// fileBytes, err := ioutil.ReadFile("inventory.json")
	// if err != nil {
	// 	fmt.Printf("Error opening file inventory.json: %e", err)
	// }
	// var fileContent map[string]interface{}
	// err = json.Unmarshal(fileBytes, &fileContent)
	// if err != nil {
	// 	fmt.Printf("Error unmarshalling file inventory.json: %e", err)
	// }
	// log.Println(fileContent)
}
