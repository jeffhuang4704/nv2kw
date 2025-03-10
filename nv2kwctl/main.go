package main

import (
	"encoding/json"
	"fmt"
	"nv2kwctl/nvapis"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {

	// GenerateYaml()
	// CreateValidatingAdmissionPolicy()

	data := map[string]string{
		"bad1":      "value1*",
		"bad2":      "value2*",
		"prohibit4": "",
	}

	result, err := MapToJSONString(data)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(result) // Output: '{"bad1": "value1*", "bad2": "value2*", "prohibit4": ""}'

}

// GenerateKindCheckExpression constructs an expression checking if object.kind is in the given list.
//   - name: isType2
//     expression: 'object.kind in ["Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"] ? true: false'
//
// Usage:
// kinds := []string{"Deployment", "ReplicaSet"}
//
//	result := GenerateKindCheckExpression(kinds)
func GenerateKindCheckExpression(kinds []string) string {
	quotedKinds := make([]string, len(kinds))
	for i, kind := range kinds {
		quotedKinds[i] = fmt.Sprintf(`"%s"`, kind) // Add quotes around each kind
	}
	return fmt.Sprintf(`'object.kind in [%s] ? true: false'`, strings.Join(quotedKinds, ","))
}

// GenerateExpression generates the expected expression based on the provided field path.
//   - name: dataset3
//     expression: 'has(object.spec.jobTemplate.metadata.annotations) ? object.spec.jobTemplate.metadata.annotations: []'
//
// Usage:
// fieldPath := "object.spec.template.metadata.annotations"
//
//	result := GenerateExpression(fieldPath)
func GenerateExpression(fieldPath string) string {
	return fmt.Sprintf("'has(%s) ? %s : []'", fieldPath, fieldPath)
}

// MapToJSONString converts a map[string]string to a JSON string formatted as expected.
func MapToJSONString(input map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("'%s'", string(jsonBytes)), nil
}

func CreateValidatingAdmissionPolicy() {
	matchConstraints := []map[string]interface{}{
		{
			"apiGroups":   []interface{}{""},
			"apiVersions": []interface{}{"v1"},
			"operations":  []interface{}{"CREATE", "UPDATE"},
			"resources":   []interface{}{"pods"},
		},
		{
			"apiGroups":   []interface{}{"apps"},
			"apiVersions": []interface{}{"v1"},
			"operations":  []interface{}{"CREATE", "UPDATE"},
			"resources":   []interface{}{"deployments", "replicasets", "daemonsets", "statefulsets"},
		},
		{
			"apiGroups":   []interface{}{"batch"},
			"apiVersions": []interface{}{"v1"},
			"operations":  []interface{}{"CREATE", "UPDATE"},
			"resources":   []interface{}{"jobs", "cronjobs"},
		},
	}

	variables := []map[string]interface{}{
		{"name": "blacklist", "expression": `{"bad1": "value1*", "bad2": "value2*", "prohibit4": ""}`},
		{"name": "dataset1a", "expression": `has(object.metadata.annotations) ? object.metadata.annotations : []`},
		{"name": "dataset1b", "expression": `has(object.spec.template.metadata.annotations) ? object.spec.template.metadata.annotations : []`},
		{"name": "dataset3", "expression": `has(object.spec.jobTemplate.metadata.annotations) ? object.spec.jobTemplate.metadata.annotations: []`},
		{"name": "isType1", "expression": `object.kind in ["Pod"] ? true: false`},
		{"name": "isType2", "expression": `object.kind in ["Deployment","ReplicaSet","DaemonSet","StatefulSet","Job"] ? true: false`},
		{"name": "isType3", "expression": `object.kind in ["CronJob"] ? true: false`},
	}

	validations := []map[string]interface{}{
		{
			"expression": `
			!variables.isType1 ||
			(
				!variables.dataset1a.exists(key, key in variables.blacklist && variables.dataset1a[key].matches(variables.blacklist[key]))
			)
			`,
			"message": "operator: pod contains_any, annotations cannot use any blacklist key/value",
		},
		{
			"expression": `
			!variables.isType2 ||
			(
				(!variables.dataset1a.exists(key, key in variables.blacklist && variables.dataset1a[key].matches(variables.blacklist[key])))
				&&
				(!variables.dataset1b.exists(key, key in variables.blacklist && variables.dataset1b[key].matches(variables.blacklist[key])))
			)
			`,
			"message": "operator: deployment contains_any, annotations cannot use any blacklist key/value",
		},
		{
			"expression": `
			!variables.isType3 ||
			(
				(!variables.dataset1a.exists(key, key in variables.blacklist && variables.dataset1a[key].matches(variables.blacklist[key])))
				&&
				(!variables.dataset3.exists(key, key in variables.blacklist && variables.dataset3[key].matches(variables.blacklist[key])))
			)
			`,
			"message": "operator: cronjob contains_any, annotations cannot use any blacklist key/value",
		},
	}

	// Generate the policy
	policy, err := GenerateValidatingAdmissionPolicy("demo1", matchConstraints, variables, validations)
	if err != nil {
		panic(err)
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(policy.Object)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(yamlData))
}

// GenerateValidatingAdmissionPolicy constructs a ValidatingAdmissionPolicy object dynamically.
func GenerateValidatingAdmissionPolicy(name string, matchConstraints []map[string]interface{}, variables, validations []map[string]interface{}) (*unstructured.Unstructured, error) {
	policy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "admissionregistration.k8s.io/v1",
			"kind":       "ValidatingAdmissionPolicy",
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec": map[string]interface{}{
				"failurePolicy":    "Fail",
				"matchConstraints": map[string]interface{}{"resourceRules": matchConstraints},
				"variables":        variables,
				"validations":      validations,
			},
		},
	}
	return policy, nil
}

func Test1_ParseJSONFile() {
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
