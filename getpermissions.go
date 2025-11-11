package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ServiceData represents the structure of each JSON file
type ServiceData struct {
	Name    string   `json:"Name"`
	Actions []Action `json:"Actions"`
}

// Action represents an action within a service
type Action struct {
	Name                string      `json:"Name"`
	ActionConditionKeys []string    `json:"ActionConditionKeys"`
	Annotations         Annotations `json:"Annotations"`
	Resources           []Resource  `json:"Resources"`
}

type Resource struct {
	Name          string   `json:"Name"`
	ConditionKeys []string `json:"ConditionKeys"`
}

// Annotations contains properties of an action
type Annotations struct {
	Properties Properties `json:"Properties"`
}

// Properties contains various flags for the action
type Properties struct {
	IsList                 bool `json:"IsList"`
	IsPermissionManagement bool `json:"IsPermissionManagement"`
	IsTaggingOnly          bool `json:"IsTaggingOnly"`
	IsWrite                bool `json:"IsWrite"`
}

// ReadAndParseJSONFiles reads all JSON files from the output directory and parses them
func ReadAndParseJSONFiles() {
	outputDir := "output"

	// Get all JSON files from the output directory
	files, err := getJSONFiles(outputDir)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	fmt.Printf("Found %d JSON files in %s directory\n", len(files), outputDir)

	// Process each file
	allServices := make(map[string]ServiceData)

	for _, file := range files {
		filePath := filepath.Join(outputDir, file)
		serviceData, err := parseJSONFile(filePath)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", file, err)
			continue
		}

		allServices[serviceData.Name] = serviceData
	}

	generateTaggingActionsCsv(allServices)
}

// getJSONFiles returns all .json files in the specified directory
func getJSONFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var jsonFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}

	return jsonFiles, nil
}

// parseJSONFile reads and parses a JSON file into ServiceData struct
func parseJSONFile(filePath string) (ServiceData, error) {
	var serviceData ServiceData

	file, err := os.Open(filePath)
	if err != nil {
		return serviceData, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return serviceData, err
	}

	err = json.Unmarshal(data, &serviceData)
	if err != nil {
		return serviceData, err
	}

	return serviceData, nil
}

func arrContains(slice []string, item string) bool {

	for _, v := range slice {
		if strings.Contains(v, item) {
			return true
		}
	}
	return false
}

func generateTaggingActionsCsv(services map[string]ServiceData) {
	f, err := os.Create("tagging-actions.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write CSV header
	if err = writer.Write([]string{"Service", "Action", "Condition", "HasTagActionCondition", "ResourcesWithCondition", "IsTaggingOnly"}); err != nil {
		fmt.Printf("Error writing CSV header: %v\n", err)
		return
	}

	for serviceName, serviceData := range services {
		fmt.Printf("Processing service: %s\n", serviceName)
		for _, action := range serviceData.Actions { // actions within a service, e.g. "CreateInstance"
			// if action.Annotations.Properties.IsTaggingOnly {
			// 	fmt.Printf("Service: %s, Tagging Action: %s\n", name, action.Name)
			// }
			var affectedResources []string

			for _, resource := range action.Resources {
				if arrContains(resource.ConditionKeys, "aws:RequestTag/${TagKey}") {
					affectedResources = append(affectedResources, resource.Name)
				}
			}
			hasTagActionCondition := arrContains(action.ActionConditionKeys, "aws:RequestTag/${TagKey}")

			if len(affectedResources) > 0 || hasTagActionCondition {
				fmt.Printf("Added service: %s with more than one affected resource\n", serviceName)

				if err = writer.Write([]string{
					serviceName,
					fmt.Sprintf("%s:%s", serviceName, action.Name),
					"aws:RequestTag/${TagKey}",
					fmt.Sprintf("%t", hasTagActionCondition),
					strings.Join(affectedResources, "\n"),
					fmt.Sprintf("%t", action.Annotations.Properties.IsTaggingOnly),
				}); err != nil {
					fmt.Printf("Error writing CSV row: %v\n", err)
					return
				}
			}
		}
	}
}
