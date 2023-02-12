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
		inComment           bool
	)

	orderOfVariable := 1
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "/*") && !strings.HasSuffix(strings.TrimSpace(line), "*/") {
			inComment = true
			continue
		} else if strings.HasPrefix(strings.TrimSpace(line), "*/") {
			inComment = false
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), "//") || inComment {
			continue
		}

		if inVariableBlock && line == "}" && nestBlockDepth == 0 {
			variableBody = append(variableBody, line)
			variables[currentVariableName] = tfVariable{
				block: strings.Join(variableBody, "\n"),
				order: orderOfVariable,
			}
			orderOfVariable++
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
					order: orderOfVariable,
				}
				orderOfVariable++
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

func collectDeclaredVariablesNew(reader io.Reader, usedVars usedVariables, sync bool) (string, error) {
	var (
		currentVariableName string
		variableFileLines   []string
	)

	ls, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(ls), "\n")
	for i := 0; i < len(lines); i++ {
		match := variablePattern.FindStringSubmatch(lines[i])
		if len(match) > 0 {
			currentVariableName = match[1]
			if _, ok := usedVars[currentVariableName]; !ok {
				if sync {
					// 使われてないvariable
					if len(match) == 3 && match[2] == "}" {
						// variable "unused" {} のパターン
						continue
					}
					j := i
					for ; j < len(lines); j++ {
						if lines[j] == "}" {
							break
						}
					}
					i = j + 1
				}
			} else {
				usedVars[currentVariableName].declared = true
			}
		}
		if i >= len(lines) {
			break
		}
		variableFileLines = append(variableFileLines, lines[i])
	}
	return strings.Join(variableFileLines, "\n"), nil
}

// not implemented %{}
func collectUsedVariables(reader io.Reader) (map[string]struct{}, error) {
	usedVariables := make(map[string]struct{})
	scanner := bufio.NewScanner(reader)
	var (
		heredocMarker string
		inComment     bool
	)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "/*") && !strings.HasSuffix(strings.TrimSpace(line), "*/") {
			inComment = true
			continue
		} else if strings.HasPrefix(strings.TrimSpace(line), "*/") {
			inComment = false
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), "//") || inComment {
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
