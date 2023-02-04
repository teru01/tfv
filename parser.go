package core

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile(`^variable "(.*?)" {$`)

func collectDeclaredVariables(reader io.Reader) (map[string]string, error) {
	variables := make(map[string]string)
	scanner := bufio.NewScanner(reader)
	var (
		inVariableBlock     bool
		nestBlockDepth      int
		currentVariableName string
		variableBody        []string
	)

	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()
		if inVariableBlock && line == "}" && nestBlockDepth == 0 {
			variableBody = append(variableBody, line)
			variables[currentVariableName] = strings.Join(variableBody, "\n")
			inVariableBlock = false
			variableBody = nil
			continue
		}
		match := variablePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			inVariableBlock = true
			currentVariableName = match[1]
			variableBody = append(variableBody, line)
			continue
		}
		if inVariableBlock {
			if strings.Contains(line, "{") {
				nestBlockDepth += 1
			}
			if strings.Contains(line, "}") {
				nestBlockDepth -= 1
			}
			variableBody = append(variableBody, line)
		}
	}

	return variables, nil
}

func collectUsedVariables(reader io.Reader) (map[string]struct{}, error) {
	usedVariables := make(map[string]struct{})
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()
		_ = line
	}
	return usedVariables, nil
}
