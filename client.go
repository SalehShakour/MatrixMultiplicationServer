package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type MatrixData struct {
	MatrixA [][]float64 `json:"matrixA"`
	MatrixB [][]float64 `json:"matrixB"`
}

func main() {

	//---------------
	// Set the new number of workers
	newNumWorkers := "4"

	// Create a HTTP POST request with the new number of workers as the body
	url := "http://localhost:9090/setnumworkers"
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(newNumWorkers))
	if err != nil {
		log.Fatal(err)
	}

	// Create a HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned error: %s", resp.Status)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Print the response body
	fmt.Println(string(body))

	// Create an instance of MatrixData with some sample data
	data := MatrixData{
		MatrixA: [][]float64{
			{1, 2, 3},
			{4, 5, 6},
		},
		MatrixB: [][]float64{
			{7, 8},
			{9, 10},
			{11, 12},
		},
	}

	// Encode the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	// Create a HTTP POST request with the JSON data as the body
	url = "http://localhost:9090/mulmatrix"
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	// Set the content type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Create a HTTP client and send the request
	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the status code of the response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Server returned error: %s", resp.Status)
	}

	// Read the response body
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Decode the response body from JSON to a slice of float64
	var result [][]float64
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)

	}

	// Print the result
	fmt.Println("The result matrix is:")
	for _, row := range result {
		fmt.Println(row)
	}
}
