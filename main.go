package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . [getfiles|parsefiles]")
		fmt.Println("  getfiles   - Download service definition files")
		fmt.Println("  parsefiles - Parse existing JSON files")
		return
	}

	command := os.Args[1]

	switch command {
	case "getfiles":
		fmt.Println("Running file download...")
		DownloadServiceFiles()
	case "parsefiles":
		fmt.Println("Parsing JSON files...")
		ReadAndParseJSONFiles()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Use 'getfiles' or 'parsefiles'")
	}
}
