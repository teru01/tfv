package core

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	variablePattern        = regexp.MustCompile(`^variable "(.*?)" {(}?)`)
	usedVarPattern         = regexp.MustCompile(`([^\w\.]|^)var\.([\w-]*)`)
	quotePattern           = regexp.MustCompile(`"(.*?[^\\])"`)
	quoteSeparationPattern = regexp.MustCompile(`".*?[^\\]"|[^"\s]+`)
	varInQuotePattern      = regexp.MustCompile(`[^$]\${var\.([\w-]*).*?}`)
	heredocPattern         = regexp.MustCompile(`<<-?([^"]*)`)
)

func collectDeclaredVariables(reader io.Reader) (tfVariables, error) {
	variables := make(tfVariables)
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
			variables[currentVariableName] = tfVariable{
				block: strings.Join(variableBody, "\n"),
			}
			inVariableBlock = false
			variableBody = nil
			continue
		}
		match := variablePattern.FindStringSubmatch(line)
		if len(match) > 0 {
			currentVariableName = match[1]
			if len(match) == 3 && match[2] == "}" {
				variables[currentVariableName] = tfVariable{
					block: line,
				}
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

// not implemented %{}, and /* */
func collectUsedVariables(reader io.Reader) (map[string]struct{}, error) {
	usedVariables := make(map[string]struct{})
	scanner := bufio.NewScanner(reader)
	var heredocMarker string
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		if heredocMarker != "" && strings.TrimSpace(line) == heredocMarker {
			heredocMarker = ""
			continue
		}
		heredocMatches := heredocPattern.FindStringSubmatch(line)
		if len(heredocMatches) > 0 {
			// start heredoc
			heredocMarker = heredocMatches[1]
			continue
		}
		if heredocMarker != "" {
			matchesInQuote := varInQuotePattern.FindAllStringSubmatch(line, -1)
			for _, m := range matchesInQuote {
				usedVariables[m[1]] = struct{}{}
			}
			continue
		}

		matches := quoteSeparationPattern.FindAllString(line, -1)
		for _, match := range matches {
			if match[0] == '"' {
				// inside quote
				matchesInQuote := varInQuotePattern.FindAllStringSubmatch(match, -1)
				for _, m := range matchesInQuote {
					usedVariables[m[1]] = struct{}{}
				}
			} else {
				// outside quote
				varMatch := usedVarPattern.FindAllStringSubmatch(match, -1)
				for _, m := range varMatch {
					usedVariables[m[2]] = struct{}{}
				}
			}
		}
	}
	return usedVariables, nil
}
