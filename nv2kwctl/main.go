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
	if err := ProcessRules("./rules/nvrules1.json"); err != nil {
		log.Fatalf("Error processing rules: %v", err)
	}
}

func ProcessRules(nvRuleFile string) error {
	ruleObj, err := ParseJSONFile(nvRuleFile)
	if err != nil {
		return fmt.Errorf("failed to parse JSON file: %w", err)
	}

	for _, rule := range ruleObj.Rules {
		if !IsSupportedRule(rule) {
			continue
		}

		policies, err := GeneratePolicies(rule)
		if err != nil {
			return fmt.Errorf("failed to generate policies: %w", err)
		}

		if err := ApplyPolicies(policies, "./templates/base.yaml"); err != nil {
			return fmt.Errorf("failed to apply policies: %w", err)
		}
	}
	return nil
}

func GeneratePolicies(rule *nvapis.RESTAdmissionRule) (map[string]interface{}, error) {
	policies := make(map[string]interface{})

	for _, criterion := range rule.Criteria {
		yamlFile := fmt.Sprintf("./templates/%s_%s.yaml", criterion.Name, criterion.Op)
		yamlObj, err := readYAML(yamlFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read or parse YAML: %w", err)
		}

		u := &unstructured.Unstructured{Object: yamlObj}
		// settings, found, err := unstructured.NestedMap(u.Object, "spec", "settings")
		settings, found, err := unstructured.NestedMap(u.Object, "spec")
		if err != nil || !found {
			return nil, fmt.Errorf("failed to extract spec.settings: %w", err)
		}

		if err := UpdateExpression(settings, criterion); err != nil {
			return nil, err
		}

		policyName := fmt.Sprintf("policy_%s_%s", criterion.Name, criterion.Op)
		policies[policyName] = settings
	}

	return policies, nil
}

func UpdateExpression(settings map[string]interface{}, criterion *nvapis.RESTAdmRuleCriterion) error {
	// Navigate to settings["settings"]["variables"]
	nestedSettings, ok := settings["settings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to find 'settings' in the provided map")
	}

	// Remove unwanted keys
	delete(settings, "mutating")
	delete(settings, "rules")
	delete(settings, "backgroundAudit")

	variables, ok := nestedSettings["variables"].([]interface{})
	if !ok {
		return fmt.Errorf("unable to find 'variables' in 'settings'")
	}

	// Iterate through variables and update
	for i, v := range variables {
		if variable, ok := v.(map[string]interface{}); ok && variable["name"] == "blacklist" {
			variable["expression"] = parseAndFormat(criterion.Value)
			variables[i] = variable
			nestedSettings["variables"] = variables
			return nil
		}
	}

	return nil
}

func ApplyPolicies(policies map[string]interface{}, baseFilePath string) error {
	baseYAML, err := readYAML(baseFilePath)
	if err != nil {
		return err
	}

	base := &unstructured.Unstructured{Object: baseYAML}
	if err := unstructured.SetNestedField(base.Object, policies, "spec", "policies"); err != nil {
		return fmt.Errorf("failed to insert policies: %w", err)
	}

	if err := unstructured.SetNestedField(base.Object, getPolicyKeys(policies), "spec", "expression"); err != nil {
		return fmt.Errorf("failed to update expression: %w", err)
	}

	finalYAML, err := yaml.Marshal(base.Object)
	if err != nil {
		return fmt.Errorf("failed to marshal final YAML: %w", err)
	}

	fmt.Println(string(finalYAML))
	fmt.Println("---")
	return nil
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

func getPolicyKeys(policies map[string]interface{}) string {
	keys := make([]string, 0, len(policies))
	for key := range policies {
		keys = append(keys, key+"()")
	}
	// return strings.Join(keys, " && ")
	return strings.Join(keys, " || ")
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
