package core

import (
	"bytes"
	"fmt"
	"io"
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
			start: v.Expr.Range().Start.Line,
			end:   v.Expr.Range().End.Line,
		}
	}

	return m, nil
}

func buildTfVars(file io.Reader, keysToDelete []string) ([]string, error) {
	var copiedFile bytes.Buffer
	teeFile := io.TeeReader(file, &copiedFile)
	tfvars, err := collectDeclaredTfvars(teeFile)
	if err != nil {
		return nil, fmt.Errorf("collect tfvars file: %w", err)
	}
	for _, key := range keysToDelete {
		delete(tfvars, key)
	}

	l, err := io.ReadAll(&copiedFile)
	if err != nil {
		return nil, fmt.Errorf("read tfvar file: %w", err)
	}
	lines := strings.Split(string(l), "\n")

	var tfvarsLine []string
	for _, v := range tfvars {
		tfvarsLine = append(tfvarsLine, lines[v.start-1:v.end]...)
	}
	return tfvarsLine, nil
}
