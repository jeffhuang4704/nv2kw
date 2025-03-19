package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func readYAML(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %w", filePath, err)
	}
	var yamlObj map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML %s: %w", filePath, err)
	}
	return yamlObj, nil
}

func modifyYAML(yamlObj map[string]interface{}) map[string]interface{} {
	// Example: Add a sample field to demonstrate modification
	yamlObj["modified"] = true
	return yamlObj
}

func main() {
	if len(os.Args) < 4 {
		log.Fatalf("Usage: %s <base.yaml> <data1.yaml> <data2.yaml>", os.Args[0])
	}

	// Read and parse YAML files
	baseYAML, err := readYAML(os.Args[1])
	if err != nil {
		log.Fatalf("%v", err)
	}

	data1YAML, err := readYAML(os.Args[2])
	if err != nil {
		log.Fatalf("%v", err)
	}

	data2YAML, err := readYAML(os.Args[3])
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Modify data1 and data2
	data1YAML = modifyYAML(data1YAML)
	data2YAML = modifyYAML(data2YAML)

	// Convert to Unstructured for manipulation
	base := &unstructured.Unstructured{Object: baseYAML}

	// Insert into spec.policies with unique names
	policies := map[string]interface{}{
		"data1": data1YAML,
		"data2": data2YAML,
	}
	if err := unstructured.SetNestedField(base.Object, policies, "spec", "policies"); err != nil {
		log.Fatalf("Failed to insert policies: %v", err)
	}

	// Convert back to YAML
	finalYAML, err := yaml.Marshal(base.Object)
	if err != nil {
		log.Fatalf("Failed to marshal final YAML: %v", err)
	}

	fmt.Println(string(finalYAML))
}
