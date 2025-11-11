package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ServiceReference struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

// DownloadServiceFiles downloads service definition files from AWS
func DownloadServiceFiles() {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://servicereference.us-east-1.amazonaws.com/", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	var result []ServiceReference
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	workers := 10
	jobs := make(chan ServiceReference, 10)

	for i := 0; i < workers; i++ {
		go func() {
			for job := range jobs {
				filename := "output/" + job.Service + ".json"
				fmt.Printf("Service: %s, URL: %s\n", job.Service, job.URL)

				if _, err := os.Stat(filename); err == nil {
					fmt.Printf("File %s already exists. Skipping write.\n", filename)
					continue
				} else if !os.IsNotExist(err) {
					fmt.Printf("Error checking file: %v\n", err)
					continue
				}

				req, err := http.NewRequest("GET", job.URL, nil)
				if err != nil {
					fmt.Println("Error creating request:", err)
					return
				}

				resp, err := httpClient.Do(req)
				if err != nil {
					fmt.Println("Error making request:", err)
					return
				}
				defer resp.Body.Close()
				var result json.RawMessage
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					fmt.Println("Error decoding response:", err)
					return
				}

				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
				if err != nil {
					fmt.Printf("Failed to create file: %v\n", err)
					return
				}
				defer f.Close()

				if _, err := f.Write(result); err != nil {
					fmt.Printf("Write error: %v\n", err)
					return
				}
			}
		}()
	}

	for _, serviceRef := range result {
		//fmt.Printf("Service: %s, URL: %s\n", serviceRef.Service, serviceRef.URL)
		jobs <- serviceRef
	}
	close(jobs)
}
