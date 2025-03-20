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

func Test1_ParseJSONFile() {
	nvRuleFile := "./rules/nvrules1.json"
	ruleObj, err := ParseJSONFile(nvRuleFile)
	if err != nil {
		fmt.Printf("failed to parse JSON file: %v\n", err)
		return
	}

	for _, rule := range ruleObj.Rules {
		if !IsSupportedRule(rule) {
			continue
		}

		// Insert into spec.policies with unique names
		policies := map[string]interface{}{}

		for _, criterion := range rule.Criteria {
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

			//TODO: update operator value, use a more generic name instead of "blacklist"
			// Modify the "expression" value
			if variables, ok := settings["variables"].([]interface{}); ok {
				for i, v := range variables {
					if variable, ok := v.(map[string]interface{}); ok && variable["name"] == "blacklist" {
						userInput := parseAndFormat(criterion.Value)
						variable["expression"] = userInput // Modify the expression here
						variables[i] = variable            //update the slice.
						settings["variables"] = variables  //Update the settings map.
						break                              // Exit the loop after modification
					}
				}
			}

			policy_name := fmt.Sprintf("policy_%s_%s", criterion.Name, criterion.Op)
			policies[policy_name] = settings
		}

		//TODO: read base.yaml, combine all the policies generated from the criteria loop above
		baseYAMLFile := "./templates/base.yaml"
		baseYAML, err := readYAML(baseYAMLFile)
		if err != nil {
			log.Fatalf("%v", err)
		}

		// Convert to Unstructured for manipulation
		base := &unstructured.Unstructured{Object: baseYAML}

		if err := unstructured.SetNestedField(base.Object, policies, "spec", "policies"); err != nil {
			log.Fatalf("Failed to insert policies: %v", err)
		}

		// format the expression
		policy_expression := getPolicyKeys(policies)
		newExpression := policy_expression
		if err := unstructured.SetNestedField(base.Object, newExpression, "spec", "expression"); err != nil {
			log.Fatalf("Failed to update expression: %v", err)
		}

		// Convert back to YAML
		finalYAML, err := yaml.Marshal(base.Object)
		if err != nil {
			log.Fatalf("Failed to marshal final YAML: %v", err)
		}

		fmt.Println(string(finalYAML))
		fmt.Println("---")
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

func IsSupportedRule(rule *nvapis.RESTAdmissionRule) bool {
	supportedMatrix := map[string]map[string]bool{}
	supportedMatrix["envVars"] = map[string]bool{
		"containsAny":       true,
		"containsAll":       true,
		"notContainsAny":    true,
		"containsOtherThan": true,
	}

	supportedMatrix["labels"] = map[string]bool{
		"containsAny":       true,
		"containsAll":       false,
		"notContainsAny":    false,
		"containsOtherThan": false,
	}

	supportedMatrix["annotations"] = map[string]bool{
		"containsAny":       true,
		"containsAll":       false,
		"notContainsAny":    false,
		"containsOtherThan": false,
	}

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

func parseAndFormat(input string) string {
	pairs := strings.Split(input, ",")
	result := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		var cleanValue string
		if !strings.ContainsAny(value, "?*") {
			cleanValue = fmt.Sprintf("^%s$", value)
		} else {
			cleanValue = strings.Replace(value, ".", "\\.", -1)
			cleanValue = strings.Replace(cleanValue, "?", ".", -1)
			cleanValue = strings.Replace(cleanValue, "*", ".*", -1)
			cleanValue = fmt.Sprintf("^%s$", cleanValue)
		}

		result[key] = cleanValue
	}

	// Build JSON-like output
	var builder strings.Builder
	builder.WriteString("{")
	first := true
	for k, v := range result {
		if !first {
			builder.WriteString(", ")
		}
		first = false
		builder.WriteString(fmt.Sprintf("\"%s\": \"%s\"", k, v))
	}
	builder.WriteString("}")

	return builder.String()
}

func getPolicyKeys(policies map[string]interface{}) string {
	keys := make([]string, 0, len(policies))
	for key := range policies {
		keys = append(keys, key+"()")
	}
	return strings.Join(keys, " && ")
}
