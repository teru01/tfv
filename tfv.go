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

type tfVariable struct {
	block string
	used  bool
	order int
}

type tfVariables map[string]tfVariable

type tfvarBlock struct {
	name  string
	start int
	end   int
}

type usedVariables map[string]*usedVar
type usedVar struct {
	declared bool
}

func GenerateVariables(ctx *cli.Context) (string, string, error) {
	usedVars, err := collectAllUsedVariables(ctx.String("dir"))
	if err != nil {
		return "", "", fmt.Errorf("collect used vars: %w", err)
	}

	variables, keysToDelete, err := rebuildVariableFile(usedVars, ctx.String("dir"), ctx.Bool("sync"))
	if err != nil {
		return "", "", fmt.Errorf("build variables blocks: %w", err)
	}

	filename := ctx.String("tfvars-file")
	if filename != "" && ctx.Bool("sync") {
		file, err := os.Open(filename)
		if err != nil {
			return "", "", fmt.Errorf("open tfvars file: %w", err)
		}
		tfvars, err := buildTfVars(file, keysToDelete)
		if err != nil {
			return "", "", fmt.Errorf("create tfvar: %w", err)
		}
		return variables, tfvars, nil
	}

	return variables, "", nil
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

func collectAllUsedVariables(dir string) (usedVariables, error) {
	usedVars := make(usedVariables)
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

		vu, err := collectUsedVariables(file)
		if err != nil {
			return nil, fmt.Errorf("collect used variables: %w", err)
		}
		for k := range vu {
			usedVars[k] = &usedVar{}
		}
	}

	return usedVars, nil
}

func rebuildVariableFile(usedVars usedVariables, path string, sync bool) (string, map[string]struct{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, fmt.Errorf("open %v: %w", path, err)
	}
	defer file.Close()

	declaredVars, unusedVariables, err := rebuildDeclaredVariables(file, usedVars, sync)
	if err != nil {
		return "", nil, fmt.Errorf("collect declared variables: %w", err)
	}

	return declaredVars, unusedVariables, nil
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
