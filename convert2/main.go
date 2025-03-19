package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path-to-yaml-file>\n", os.Args[0])
	}
	yamlFile := os.Args[1]
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Convert YAML to Unstructured
	var yamlObj map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	u := &unstructured.Unstructured{Object: yamlObj}

	// Extract spec.settings
	settings, found, err := unstructured.NestedMap(u.Object, "spec", "settings")
	if err != nil || !found {
		log.Fatalf("Failed to extract spec.settings: %v", err)
	}

	// Convert to YAML
	extractedYAML, err := yaml.Marshal(map[string]interface{}{"settings": settings})
	if err != nil {
		log.Fatalf("Failed to marshal extracted settings to YAML: %v", err)
	}

	fmt.Println(string(extractedYAML))
}

