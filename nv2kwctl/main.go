package main

import (
	"encoding/json"
	"fmt"
	"log"
	"nv2kwctl/nvapis"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	Test1_ParseJSONFile()
}

// func main() {

// 	// GenerateYaml()
// 	// CreateValidatingAdmissionPolicy()

// 	data := map[string]string{
// 		"bad1":      "value1*",
// 		"bad2":      "value2*",
// 		"prohibit4": "",
// 	}

// 	result, err := MapToJSONString(data)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}

// 	fmt.Println(result) // Output: '{"bad1": "value1*", "bad2": "value2*", "prohibit4": ""}'

// }

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
		fmt.Printf("ID: %d\n", rule.ID)
		// fmt.Printf("Category: %s\n", rule.Category)
		// fmt.Printf("Comment: %s\n", rule.Comment)
		// fmt.Printf("Disable: %t\n", rule.Disable)
		// fmt.Printf("Critical: %t\n", rule.Critical)
		// fmt.Printf("CfgType: %s\n", rule.CfgType)
		// fmt.Printf("RuleType: %s\n", rule.RuleType)
		// fmt.Printf("RuleMode: %s\n", rule.RuleMode)
		// fmt.Printf("Criteria count: %d\n", len(rule.Criteria))

		//TODO: check if all the criteria in this rule are all supported
		// only namespace, envvar, labels and annotation are supported
		if !IsSupportedRule(rule) {
			// fmt.Printf("Rule is not supported\n")
			continue
		}

		sliceOfMaps := []map[string]interface{}{}
		for _, criterion := range rule.Criteria {
			fmt.Printf("	Name: %s\n", criterion.Name)
			fmt.Printf("	Op: %s\n", criterion.Op)
			fmt.Printf("	Value: %s\n", criterion.Value)
			fmt.Printf("	Type: %s\n", criterion.Type)
			fmt.Printf("	Kind: %s\n", criterion.Kind)
			fmt.Printf("	Path: %s\n", criterion.Path)
			fmt.Printf("	ValueType: %s\n", criterion.ValueType)
			fmt.Println("	=============")

			//TODO: generate a CEL expression for this criterion
			//1. use the criteria name and operator to find the corresponding yaml file
			//2. read the yaml file
			//3. convert the criterion.Value to the corresponding value in the yaml file
			//4. store the  converted policy (for kw)

			yamlFile := fmt.Sprintf("./templates/%s_%s.yaml", criterion.Name, criterion.Op)

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

			// add to the slice of maps
			sliceOfMaps = append(sliceOfMaps, settings)

			// Convert to YAML
			// extractedYAML, err := yaml.Marshal(map[string]interface{}{"settings": settings})
			// if err != nil {
			// 	log.Fatalf("Failed to marshal extracted settings to YAML: %v", err)
			// }

			// fmt.Println(string(extractedYAML))
		}

		//TODO: read base.yaml
		// combine all the policies generated from the criteria loop above

		// Insert into spec.policies with unique names
		// policies := map[string]interface{}{
		// 	"data1": data1YAML,
		// 	"data2": data2YAML,
		// }
		// if err := unstructured.SetNestedField(base.Object, policies, "spec", "policies"); err != nil {
		// 	log.Fatalf("Failed to insert policies: %v", err)
		// }

		baseYAMLFile := "./templates/base.yaml"
		baseYAML, err := readYAML(baseYAMLFile)
		if err != nil {
			log.Fatalf("%v", err)
		}

		// Convert to Unstructured for manipulation
		base := &unstructured.Unstructured{Object: baseYAML}

		// Insert into spec.policies with unique names
		policies := map[string]interface{}{
			// "data1": data1YAML,
			// "data2": data2YAML,
		}

		// Loop through the sliceOfMaps
		for index, mapData := range sliceOfMaps {
			// fmt.Printf("Map %d:\n", index) // Print the index of the map

			// Loop through the key-value pairs of each map
			// for key, value := range mapData {
			// 	fmt.Printf("  %s: %v (%T)\n", key, value, value) // Print key, value, and type
			// }
			// fmt.Println() // Add a newline for readability
			keyName := fmt.Sprintf("policy_%d", index)
			policies[keyName] = mapData
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

		fmt.Println()
	}
}

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

// 10 const (
// 	9         CriteriaOpEqual               string = "="
// 	8         CriteriaOpNotEqual            string = "!="
// 	7         CriteriaOpContains            string = "contains"
// 	6         CriteriaOpPrefix              string = "prefix"
// 	5         CriteriaOpRegex               string = "regex"
// 	4         CriteriaOpNotRegex            string = "!regex"
// 	3         CriteriaOpBiggerEqualThan     string = ">="
// 	2         CriteriaOpBiggerThan          string = ">"
// 	1         CriteriaOpLessEqualThan       string = "<="
// 	0         CriteriaOpContainsAll         string = "containsAll"
// 	1         CriteriaOpContainsAny         string = "containsAny"
// 	2         CriteriaOpNotContainsAny      string = "notContainsAny"
// 	3         CriteriaOpContainsOtherThan   string = "containsOtherThan"
// 	4         CriteriaOpRegexContainsAny    string = "regexContainsAnyEx"
// 	5         CriteriaOpRegexNotContainsAny string = "!regexContainsAnyEx"
// 	6         CriteriaOpExist               string = "exist"
// 	7         CriteriaOpNotExist            string = "notExist"
// 	8         CriteriaOpContainsTagAny      string = "containsTagAny"
// 	9
//    10         CriteriaOpRegex_Deprecated    string = "regexContainsAny"  // notice: it's the same as CriteriaOpRegex since 5.3.2
//    11         CriteriaOpNotRegex_Deprecated string = "!regexContainsAny" // notice: it's the same as CriteriaOpNotRegex since 5.3.2
//    12 )
//    13

func IsSupportedRule(rule *nvapis.RESTAdmissionRule) bool {
	supportedMatrix := map[string]map[string]bool{}

	//TODO: add more supported criteria
	//TODO: fill the right operators for each criteria
	// supportedMatrix["namespace"] = map[string]bool{
	// 	"containsAny": true,
	// 	"containsAll": true,
	// }

	supportedMatrix["envVars"] = map[string]bool{
		"containsAny":       true,
		"containsAll":       true,
		"notContainsAny":    true,
		"containsOtherThan": true,
	}

	// supportedMatrix["labels"] = map[string]bool{
	// 	"containsAny":       true,
	// 	"containsAll":       true,
	// 	"notContainsAny":    true,
	// 	"containsOtherThan": true,
	// }

	// supportedMatrix["annotations"] = map[string]bool{
	// 	"containsAny":       true,
	// 	"containsAll":       true,
	// 	"notContainsAny":    true,
	// 	"containsOtherThan": true,
	// }

	for _, c := range rule.Criteria {
		if suppored, ok := supportedMatrix[c.Name][c.Op]; !ok {
			return false
		} else if !suppored {
			return false
		}
	}
	return true
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
