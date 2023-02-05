package core

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

var (
	variablePattern   = regexp.MustCompile(`^variable "(.*?)" {(}?)`)
	usedVarPattern    = regexp.MustCompile(`[^\w]var\.([\w-]*)`)
	quotePattern      = regexp.MustCompile(`"(.*?[^\\])"`)
	varInQuotePattern = regexp.MustCompile(`[^$]\${var\.([\w-]*).*?}`)
	heredocPattern    = regexp.MustCompile(`<<-?([^"]*)`)
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

// match: var.hoge var.hoge[0], var.hoge.foo, "++${var.goo}+++, "++%{ if var.name == }++"
// not match: "http://var.hoge.com", "+$${var.hoge}+"
// <<EOF, <<-EOF
// EOF

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
		} else if heredocMarker != "" {
			// inside heredoc
			continue
		}
		heredocMatches := heredocPattern.FindStringSubmatch(line)
		if len(heredocMatches) > 0 {
			heredocMarker = heredocMatches[1]
			// start heredoc
			continue
		}

		matches := usedVarPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		for _, match := range matches {
			usedVariables[match[1]] = struct{}{}
		}
		quoteMatches := quotePattern.FindAllStringSubmatch(line, -1)

		notVarInQuote := make(map[string]struct{})
		for _, qMatch := range quoteMatches {
			varMatch := usedVarPattern.FindAllStringSubmatch(qMatch[1], -1)
			for _, vm := range varMatch {
				notVarInQuote[vm[1]] = struct{}{}
			}

			matchesInQuote := varInQuotePattern.FindAllStringSubmatch(qMatch[1], -1)
			for _, m := range matchesInQuote {
				delete(notVarInQuote, m[1])
			}
		}
		for forDelete := range notVarInQuote {
			delete(usedVariables, forDelete)
		}
	}
	return usedVariables, nil
}
