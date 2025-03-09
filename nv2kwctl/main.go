package main

import (
	"encoding/json"
	"fmt"
	"nv2kwctl/nvapis"
	"os"
)

func main() {
	nvRuleFile := "./rules/nvrules1.json"
	ruleObj, err := ParseJSONFile(nvRuleFile)
	if err != nil {
		fmt.Printf("failed to parse JSON file: %v\n", err)
		return
	}
	fmt.Printf("%v\n", ruleObj)
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
