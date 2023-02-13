package core

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

func collectDeclaredTfvars(reader io.Reader) (map[string]*tfvarBlock, error) {
	config := make(map[string]*hcl.Attribute)
	src, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	if err := hclsimple.Decode("variables.hcl", src, nil, &config); err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}

	m := make(map[string]*tfvarBlock)
	for k, v := range config {
		m[k] = &tfvarBlock{
			name:  k,
			start: v.Expr.Range().Start.Line,
			end:   v.Expr.Range().End.Line,
		}
	}

	return m, nil
}

func buildTfVars(file io.Reader, keysToDelete map[string]struct{}) (string, error) {
	var copiedFile bytes.Buffer
	teeFile := io.TeeReader(file, &copiedFile)
	tfvars, err := collectDeclaredTfvars(teeFile)
	if err != nil {
		return "", fmt.Errorf("collect tfvars file: %w", err)
	}
	l, err := io.ReadAll(&copiedFile)
	if err != nil {
		return "", fmt.Errorf("read tfvar file: %w", err)
	}
	lines := strings.Split(string(l), "\n")

	var (
		tfvarsLine []string
		tfVarsList []*tfvarBlock
	)
	for _, v := range tfvars {
		tfVarsList = append(tfVarsList, v)
	}
	// 定義順序を復元するためにソート
	sort.Slice(tfVarsList, func(i, j int) bool {
		return tfVarsList[i].start < tfVarsList[j].start
	})

	varsI := 0
	for i := 0; i < len(lines); i++ {
		if varsI < len(tfVarsList) && i == tfVarsList[varsI].start-1 {
			v := tfVarsList[varsI]
			varsI++
			if _, ok := keysToDelete[v.name]; ok {
				i = v.end - 1
				continue
			}
		}
		tfvarsLine = append(tfvarsLine, lines[i])
	}

	return strings.Join(tfvarsLine, "\n"), nil
}
