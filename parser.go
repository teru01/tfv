package core

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

var (
	variablePattern        = regexp.MustCompile(`^variable "(.*?)" {(}?)`)
	usedVarPattern         = regexp.MustCompile(`([^\w\.]|^)var\.([\w-]*)`)
	quoteSeparationPattern = regexp.MustCompile(`".*?[^\\]"|[^"\s]+`)
	varInQuotePattern      = regexp.MustCompile(`[^$]\${var\.([\w-]*).*?}`)
	heredocPattern         = regexp.MustCompile(`<<-?([^"]*)`)
	outputTemplate         = `variable "%s" {
  description = ""
}
`
)

func rebuildDeclaredVariables(reader io.Reader, usedVars usedVariables, sync bool) (string, map[string]struct{}, error) {
	var (
		currentVariableName string
		variableFileLines   []string
	)
	unusedVariables := make(map[string]struct{})

	if reader != nil {
		ls, err := io.ReadAll(reader)
		if err != nil {
			return "", nil, err
		}
		lines := strings.Split(string(ls), "\n")
		for i := 0; i < len(lines); i++ {
			match := variablePattern.FindStringSubmatch(lines[i])
			if len(match) == 0 {
				if !(len(variableFileLines) == 0 && lines[i] == "") {
					// ファイル先頭の空行は読み飛ばす
					variableFileLines = append(variableFileLines, lines[i])
				}
				continue
			}
			currentVariableName = match[1]
			if _, ok := usedVars[currentVariableName]; !ok {
				// 使われてないvariable
				unusedVariables[currentVariableName] = struct{}{}
				if !sync {
					variableFileLines = append(variableFileLines, lines[i])
					continue
				}
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

				// 読み飛ばし後の空行も不要なので読み飛ばす
				for ; i < len(lines); i++ {
					if lines[i] != "" {
						// 1行読み戻す
						i--
						break
					}
				}
			} else {
				usedVars[currentVariableName].declared = true
				if !(len(variableFileLines) == 0 && lines[i] == "") {
					// ファイル先頭の空行は読み飛ばす
					variableFileLines = append(variableFileLines, lines[i])
				}
			}
		}
	}

	unDeclaredVarsLine := generateUndeclaredVariables(usedVars)
	variableFileLines = append(variableFileLines, unDeclaredVarsLine...)
	return strings.Join(variableFileLines, "\n"), unusedVariables, nil
}

func generateUndeclaredVariables(usedVars usedVariables) []string {
	unDeclaredVarsLine := make([]string, 0, len(usedVars))
	for k, v := range usedVars {
		if !v.declared {
			unDeclaredVarsLine = append(unDeclaredVarsLine, fmt.Sprintf(outputTemplate, k))
		}
	}
	sort.Slice(unDeclaredVarsLine, func(i, j int) bool {
		return unDeclaredVarsLine[i] < unDeclaredVarsLine[j]
	})
	return unDeclaredVarsLine
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
