package sender

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GraphQL endpoint URL
const url = "http://example.com/graphql"

func GraphQL(jsonInput []byte) {
	// Read JSON file
	// jsonFile, err := ioutil.ReadFile(file)
	// if err != nil {
	// 	fmt.Println("Failed to read JSON file:", err)
	// 	return
	// }

	// // Parse JSON data
	// var jsonData interface{}
	// err = json.Unmarshal(jsonFile, &jsonData)
	// if err != nil {
	// 	fmt.Println("Failed to parse JSON data:", err)
	// 	return
	// }

	// // Convert JSON data to GraphQL input format
	// jsonInput, err := json.Marshal(jsonData)
	// if err != nil {
	// 	fmt.Println("Failed to marshal JSON input:", err)
	// 	return
	// }

	// Create HTTP request with GraphQL input as payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonInput))
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		return
	}

	// Set GraphQL content type header
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return
	}
	defer res.Body.Close()

	// Read response from GraphQL endpoint
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return
	}

	// Print response
	fmt.Println(string(responseBody))
}
