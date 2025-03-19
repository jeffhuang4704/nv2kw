package main

import (
	"encoding/json"
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

	// Convert YAML to JSON (unstructured works well with JSON)
	jsonData, err := yamlToJSON(data)
	if err != nil {
		log.Fatalf("Failed to convert YAML to JSON: %v", err)
	}

	// Convert JSON to Unstructured
	var unstrObj map[string]interface{}
	if err := json.Unmarshal(jsonData, &unstrObj); err != nil {
		log.Fatalf("Failed to unmarshal to map: %v", err)
	}

	u := &unstructured.Unstructured{Object: unstrObj}

	// Manipulate the data
	updateBlacklist(u)

	// Convert back to YAML with 2-space indent
	updatedData, err := unstructuredToYAML(u)
	if err != nil {
		log.Fatalf("Failed to convert to YAML: %v", err)
	}

	fmt.Println(string(updatedData))
}

func yamlToJSON(data []byte) ([]byte, error) {
	var obj interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func unstructuredToYAML(u *unstructured.Unstructured) ([]byte, error) {
	jsonData, err := u.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var yamlObj interface{}
	if err := json.Unmarshal(jsonData, &yamlObj); err != nil {
		return nil, err
	}

	// Convert to YAML with 2-space indentation using yaml.Encoder
	buf := &yaml.Node{}
	if err := buf.Encode(yamlObj); err != nil {
		return nil, err
	}

	out := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			buf,
		},
	}

	var result []byte
	writer := yaml.NewEncoder(&sliceWriter{&result})
	writer.SetIndent(2) // Set to 2-space indentation
	defer writer.Close()

	if err := writer.Encode(out); err != nil {
		return nil, err
	}
	return result, nil
}

type sliceWriter struct {
	data *[]byte
}

func (w *sliceWriter) Write(p []byte) (int, error) {
	*w.data = append(*w.data, p...)
	return len(p), nil
}

func updateBlacklist(u *unstructured.Unstructured) {
	path := []string{"spec", "settings", "variables"}
	variables, found, err := unstructured.NestedSlice(u.Object, path...)
	if err != nil || !found {
		log.Fatalf("Failed to find variables: %v", err)
	}

	// Modify blacklist
	for i, v := range variables {
		variable := v.(map[string]interface{})
		if variable["name"] == "blacklist" {
			variable["expression"] = "{\"newbad1\": \"^newvalue1$\", \"newbad2\": \"^newvalue2$\"}"
			variables[i] = variable
		}
	}
	_ = unstructured.SetNestedSlice(u.Object, variables, path...)
}

// example code
// extract portion of the value
func Extract_portion() {
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
