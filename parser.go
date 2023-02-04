package core

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	variablePattern = regexp.MustCompile(`^variable "(.*?)" {(}?)`)
	usedVarPattern  = regexp.MustCompile(`var\.([\w-]*)`)
)

// match: var.hoge var.hoge[0], var.hoge.foo, "++${var.goo}+++, "++%{ if var.name == }++"
// not match: "http://var.hoge.com", "+$${var.hoge}+"
// <<EOF, <<-EOF
// EOF

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
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if inVariableBlock && line == "}" && nestBlockDepth == 0 {
			variableBody = append(variableBody, line)
			variables[currentVariableName] = strings.Join(variableBody, "\n")
			inVariableBlock = false
			variableBody = nil
			continue
		}
		match := variablePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			currentVariableName = match[1]
			if len(match) == 3 && match[2] == "}" {
				variables[currentVariableName] = line
				continue
			}
			variableBody = append(variableBody, line)
			inVariableBlock = true
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
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		matches := usedVarPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			usedVariables[match[1]] = struct{}{}
		}
	}
	return usedVariables, nil
}
