package main

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type Example struct {
	Name  string `yaml:"name"`
	Value int    `yaml:"value"`
}

func main() {

	// Task: Read YAML file
	// yamlData, err := readYAML("base.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("%+v\n", yamlData)

	// Task: convert struct to yaml
	// data := Example{Name: "test", Value: 1}
	// yamlString, err := structToYAML(data)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(yamlString)

	// Task: updateNestedField
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// err = updateNestedField(yamlData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// Task: updateArray
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if err := updateArray2(yamlData, "metadata.level.testdata", "newvalue2"); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Printf("%+v\n", yamlData)

	// Task: delete nested field
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// deleteNestedField(yamlData)
	// fmt.Println("=====================================")
	// printYaml(yamlData)

	// Task: add environment variable into containers[0].env
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// newEnv := map[string]interface{}{
	// 	"name":  "key3",
	// 	"value": "value3",
	// }

	// if err := addEnvVariable(yamlData, newEnv); err != nil {
	// 	fmt.Println("Failed to add environment variable:", err)
	// 	return
	// }
	// fmt.Println("=====================================")
	// printYaml(yamlData)

	//
	// Task: add environment variable into containers[0].env
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// newEnv := map[string]interface{}{
	// 	"name":  "key3",
	// 	"value": "value3",
	// }
	// addEnvVariablev2(yamlData, "nginx", newEnv)

	// fmt.Println("=====================================")
	// printYaml(yamlData)

	// Task: replace Items
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// newItems := []map[string]interface{}{
	// 	{"name": "newitem111", "value": "newvalue111"},
	// }

	// replaceItems(yamlData, newItems)

	// fmt.Println("=====================================")
	// printYaml(yamlData)

	// Task: set a map
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// updateNestedMap(yamlData)

	// fmt.Println("=====================================")
	// printYaml(yamlData)

	// Task: update string slice
	// yamlData, err := readYAML("envVars_containsAny.yaml")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// printYaml(yamlData)

	// updateStringSlice(yamlData)

	// fmt.Println("=====================================")
	// printYaml(yamlData)

	// Task: get settings
	yamlData, err := readYAML("envVars_containsAny.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	settings, err := getSettings(yamlData)
	if err != nil {
		fmt.Println(err)
		return
	}
	printYaml(settings)
}

func getSettings(yamlData map[string]interface{}) (map[string]interface{}, error) {
	// Get the settings field
	settings, found, err := unstructured.NestedMap(yamlData, "spec", "settings")
	if err != nil || !found {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	return settings, nil
}

func updateStringSlice(yamlData map[string]interface{}) error {
	newTags := []string{"tag1", "tag2", "tag3"}

	// TODO: survey this ==> unstructured.SetNestedStringMap()
	if err := unstructured.SetNestedStringSlice(yamlData, newTags, "spec", "tags"); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}
	return nil
}

func updateNestedMap(yamlData map[string]interface{}) error {
	newConfig := map[string]interface{}{
		"timeout":  "30s",
		"retries":  int64(3), // *** using int64 is important !!!
		"logLevel": "debug",
	}

	if err := unstructured.SetNestedMap(yamlData, newConfig, "spec", "settings", "config"); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}
	return nil
}

func replaceItems(yamlData map[string]interface{}, newItems []map[string]interface{}) error {
	// Convert []map[string]interface{} to []interface{}
	var items []interface{}
	for _, item := range newItems {
		items = append(items, item)
	}

	// Set the new items to spec.settings.items
	path := []string{"spec", "settings", "items"}
	if err := unstructured.SetNestedSlice(yamlData, items, path...); err != nil {
		return fmt.Errorf("failed to replace items: %w", err)
	}
	return nil
}

func printYaml(yamlData map[string]interface{}) error {
	finalYAML, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal final YAML: %w", err)
	}

	fmt.Println(string(finalYAML))
	return nil
}

func deleteNestedField(yamlData map[string]interface{}) error {
	// Remove the nested field using RemoveNestedField
	unstructured.RemoveNestedField(yamlData, "metadata", "level", "testdata")
	return nil
}

// level:
//
//	another: info
//	testdata: null
func deleteNestedField_set_null(yamlData map[string]interface{}) error {
	if err := unstructured.SetNestedField(yamlData, nil, "metadata", "level", "testdata"); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}
	return nil
}

func updateArray(yamlData map[string]interface{}, key string, newValue interface{}) error {
	arr, found, err := unstructured.NestedSlice(yamlData, key)
	if err != nil || !found {
		return fmt.Errorf("failed to get nested field: %w", err)
	}

	arr = append(arr, newValue)

	if err := unstructured.SetNestedSlice(yamlData, arr, key); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}
	return nil
}

// Usage:
// newEnv := map[string]interface{}{
// 	"name":  "key3",
// 	"value": "value3",
// }

// if err := addEnvVariable(yamlData, newEnv); err != nil {
// 	fmt.Println("Failed to add environment variable:", err)
// 	return
// }

func addEnvVariable(yamlData map[string]interface{}, newEnv map[string]interface{}) error {
	// Path to containers[0].env
	path := []string{"spec", "template", "spec", "containers"}

	// Get the containers array
	containers, found, err := unstructured.NestedSlice(yamlData, path...)
	if err != nil || !found || len(containers) == 0 {
		return fmt.Errorf("failed to get containers: %v", err)
	}

	// Access the first container (containers[0])
	container := containers[0].(map[string]interface{})

	// Get the env array
	env, found, err := unstructured.NestedSlice(container, "env")
	if err != nil {
		return fmt.Errorf("failed to get env: %v", err)
	}
	if !found {
		// If env doesn't exist, create an empty array
		env = []interface{}{}
	}

	// Append the new environment variable
	env = append(env, newEnv)

	// Set the updated env back to the container
	if err := unstructured.SetNestedSlice(container, env, "env"); err != nil {
		return fmt.Errorf("failed to set env: %v", err)
	}

	// Update the containers array
	containers[0] = container
	if err := unstructured.SetNestedSlice(yamlData, containers, path...); err != nil {
		return fmt.Errorf("failed to update containers: %v", err)
	}

	return nil
}

// addEnvVariable finds a container by name and adds an env variable to it
func addEnvVariablev2(yamlData map[string]interface{}, containerName string, newEnv map[string]interface{}) error {
	// Path to containers
	path := []string{"spec", "template", "spec", "containers"}

	// Get the containers array
	containers, found, err := unstructured.NestedSlice(yamlData, path...)
	if err != nil || !found {
		return fmt.Errorf("failed to get containers: %v", err)
	}

	// Traverse to find the target container by name
	var containerIndex int = -1
	for i, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if name, _, _ := unstructured.NestedString(container, "name"); name == containerName {
			containerIndex = i
			break
		}
	}

	if containerIndex == -1 {
		return fmt.Errorf("container with name %s not found", containerName)
	}

	container := containers[containerIndex].(map[string]interface{})

	// Get or initialize the env array
	env, found, err := unstructured.NestedSlice(container, "env")
	if err != nil {
		return fmt.Errorf("failed to get env: %v", err)
	}
	if !found {
		env = []interface{}{}
	}

	// Append the new environment variable
	env = append(env, newEnv)

	// Set the updated env back to the container
	if err := unstructured.SetNestedSlice(container, env, "env"); err != nil {
		return fmt.Errorf("failed to set env: %v", err)
	}

	// Update the container in the containers array
	containers[containerIndex] = container
	if err := unstructured.SetNestedSlice(yamlData, containers, path...); err != nil {
		return fmt.Errorf("failed to update containers: %v", err)
	}

	return nil
}

// support nested keys
func updateArray2(yamlData map[string]interface{}, key string, newValue interface{}) error {
	// Split the key by "." to support nested keys
	keys := strings.Split(key, ".")

	// Get the nested slice
	arr, found, err := unstructured.NestedSlice(yamlData, keys...)
	if err != nil || !found {
		return fmt.Errorf("failed to get nested field: %w", err)
	}

	// Append the new value
	arr = append(arr, newValue)

	// Set the updated slice back into the map
	if err := unstructured.SetNestedSlice(yamlData, arr, keys...); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}
	return nil
}

func updateNestedField(yamlData map[string]interface{}) error {
	// Check if "metadata" exists and initialize it if necessary
	if _, ok := yamlData["metadata"]; !ok {
		yamlData["metadata"] = make(map[string]interface{})
	}

	// Set a nested field
	if err := unstructured.SetNestedField(yamlData, "newValue111", "metadata", "name"); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}

	// Get a nested field
	value, found, err := unstructured.NestedString(yamlData, "metadata", "name")
	if err != nil || !found {
		return fmt.Errorf("failed to get nested field: %w", err)
	}
	fmt.Println("Field value:", value)
	return nil
}

func updateNestedField22(yamlData map[string]interface{}) error {
	u := &unstructured.Unstructured{Object: yamlData}

	// Set a nested field
	if err := unstructured.SetNestedField(u.Object, "newValue", "metadata", "name"); err != nil {
		return fmt.Errorf("failed to set nested field: %w", err)
	}

	// Get a nested field
	value, found, err := unstructured.NestedString(u.Object, "metadata", "name")
	if err != nil || !found {
		return fmt.Errorf("failed to get nested field: %w", err)
	}
	fmt.Println("Field value:", value)
	return nil
}

func structToYAML(obj interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), nil
}

func readYAML1(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return yamlData, nil
}

func readYAML(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Debugging: Print the unmarshaled YAML data to see its structure
	// fmt.Println("Unmarshaled YAML Data:", yamlData)

	return yamlData, nil
}
