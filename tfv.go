package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

var outputTemplate = `variable "%s" {
  description = ""
}`

type tfVariable struct {
	block string
	used  bool
	order int
}

type tfVariables map[string]tfVariable

type tfvarBlock struct {
	start int
	end   int
}

func GenerateVariables(ctx *cli.Context) (string, string, error) {
	variableBlocks, err := buildVariableBlocks(ctx.String("dir"))
	if err != nil {
		return "", "", fmt.Errorf("build variables blocks: %w", err)
	}

	var keysToDelete []string
	if ctx.Bool("sync") {
		for k, v := range variableBlocks {
			if !v.used {
				keysToDelete = append(keysToDelete, k)
			}
		}
		for _, key := range keysToDelete {
			delete(variableBlocks, key)
		}
	}

	var tfvarsLine []string
	filename := ctx.String("tfvars-file")
	if filename != "" && ctx.Bool("sync") {
		file, err := os.Open(filename)
		if err != nil {
			return "", "", fmt.Errorf("open tfvars file: %w", err)
		}
		tfvarsLine, err = buildTfVars(file, keysToDelete)
		if err != nil {
			return "", "", fmt.Errorf("create tfvar: %w", err)
		}
	}

	return buildVariableString(variableBlocks), strings.Join(tfvarsLine, "\n"), nil
}

func buildVariableString(vars tfVariables) string {
	var ov []tfVariable
	var ovs []string
	for _, v := range vars {
		ov = append(ov, v)
	}
	sort.Slice(ov, func(i, j int) bool {
		return ov[i].order < ov[j].order
	})
	for _, v := range ov {
		ovs = append(ovs, v.block)
	}
	return strings.Join(ovs, "\n\n")
}

func buildVariableBlocks(dir string) (tfVariables, error) {
	variableBlocks := make(tfVariables)
	usedVars := make(map[string]struct{})

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("path %v: %w", dir, err)
	}
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() || !strings.HasSuffix(path, ".tf") {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %v: %w", path, err)
		}
		defer file.Close()

		var copiedFile bytes.Buffer
		teeFile := io.TeeReader(file, &copiedFile)

		declaredVars, err := collectDeclaredVariables(teeFile)
		if err != nil {
			return nil, fmt.Errorf("collect declared variables: %w", err)
		}
		for k, v := range declaredVars {
			variableBlocks[k] = v
		}

		vu, err := collectUsedVariables(&copiedFile)
		if err != nil {
			return nil, fmt.Errorf("collect used variables: %w", err)
		}
		for k := range vu {
			usedVars[k] = struct{}{}
		}
	}

	for used := range usedVars {
		variable, ok := variableBlocks[used]
		if ok {
			variableBlocks[used] = tfVariable{
				block: variable.block,
				used:  true,
				order: variable.order,
			}
		} else {
			variableBlocks[used] = tfVariable{
				block: fmt.Sprintf(outputTemplate, used),
				used:  true,
				order: variable.order,
			}
		}
	}
	return variableBlocks, nil
}
