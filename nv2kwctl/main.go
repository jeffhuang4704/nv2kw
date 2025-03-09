package main

import (
	"encoding/json"
	"fmt"
	"nv2kwctl/nvapis"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	nvRuleFile := "./rules/nvrules1.json"
	ruleObj, err := ParseJSONFile(nvRuleFile)
	if err != nil {
		fmt.Printf("failed to parse JSON file: %v\n", err)
		return
	}
	fmt.Printf("%v\n", ruleObj)
	for _, rule := range ruleObj.Rules {
		// print
		fmt.Printf("ID: %d\n", rule.ID)
		fmt.Printf("Category: %s\n", rule.Category)
		fmt.Printf("Comment: %s\n", rule.Comment)
		fmt.Printf("Disable: %t\n", rule.Disable)
		fmt.Printf("Critical: %t\n", rule.Critical)
		fmt.Printf("CfgType: %s\n", rule.CfgType)
		fmt.Printf("RuleType: %s\n", rule.RuleType)
		fmt.Printf("RuleMode: %s\n", rule.RuleMode)
		fmt.Printf("Criteria count: %d\n", len(rule.Criteria))

		for _, criterion := range rule.Criteria {
			fmt.Printf("	Name: %s\n", criterion.Name)
			fmt.Printf("	Op: %s\n", criterion.Op)
			fmt.Printf("	Value: %s\n", criterion.Value)
			fmt.Printf("	Type: %s\n", criterion.Type)
			fmt.Printf("	Kind: %s\n", criterion.Kind)
			fmt.Printf("	Path: %s\n", criterion.Path)
			fmt.Printf("	ValueType: %s\n", criterion.ValueType)
			fmt.Println("	=============")
		}
		fmt.Println()
	}

	// GenerateYaml()
}

// ParseJSONFile reads a JSON file and unmarshals it into RESTAdmissionRulesData
func ParseJSONFile(filename string) (*nvapis.RESTAdmissionRulesData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var data nvapis.RESTAdmissionRulesData
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &data, nil
}

// This code is used in Kubernetes client-go to manage arbitrary Kubernetes objects without defining strict Go structs.
func GenerateYaml() {
	// Create an unstructured object
	policy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "policies.kubewarden.io/v1",
			"kind":       "ClusterAdmissionPolicy",
			"metadata": map[string]interface{}{
				"name": "celtest1",
				"annotations": map[string]interface{}{
					"io.kubewarden.policy.category": "Resource validation",
					"io.kubewarden.policy.severity": "low",
				},
			},
			"spec": map[string]interface{}{
				"module": "registry://ghcr.io/kubewarden/policies/cel-policy:latest",
				"settings": map[string]interface{}{
					"variables": []interface{}{
						map[string]interface{}{
							"name":       "replicas",
							"expression": "object.spec.replicas",
						},
					},
					"validations": []interface{}{
						map[string]interface{}{
							"expression": "variables.replicas <= 15",
							"message":    "The number of replicas must be less than or equal to 15",
						},
					},
				},
				"rules": []interface{}{
					map[string]interface{}{
						"apiGroups":   []interface{}{"apps"},
						"apiVersions": []interface{}{"v1"},
						"operations":  []interface{}{"CREATE", "UPDATE"},
						"resources":   []interface{}{"deployments"},
					},
				},
				"mutating":        false,
				"backgroundAudit": false,
			},
		},
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(policy.Object)
	if err != nil {
		fmt.Println("Error marshalling to YAML:", err)
		return
	}

	// Print YAML output
	fmt.Println(string(yamlData))
}
