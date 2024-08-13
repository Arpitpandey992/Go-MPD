package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func SearchWithUserInput(audioMeilisearchClient *AudioMeilisearchClient) {
	scanner := bufio.NewScanner(os.Stdin) // Create a new scanner to read from standard input

	fmt.Println("Enter text (type 'exit' to quit):")

	for {
		// Read input from the user
		scanner.Scan()
		input := scanner.Text()

		// Check if input is 'exit' to break the loop
		if strings.TrimSpace(input) == "exit" {
			fmt.Println("Exiting...")
			break
		}
		audioMetadataList, err := audioMeilisearchClient.SearchAudioFiles(input, 10)
		if err != nil {
			fmt.Printf("error: %v", err)
			continue
		}
		fmt.Println("search results:")
		for _, audioMetadata := range audioMetadataList {
			jsonString, err := audioMetadata.ToIndentedJsonString()
			if err != nil {
				fmt.Print(err)
				continue
			}
			fmt.Println(jsonString)
		}
	}
}
