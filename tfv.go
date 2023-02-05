package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/urfave/cli/v2"
)

var outputTemplate = `variable "%s" {
  description = ""
}`

type tfVariable struct {
	block string
	used  bool
}

type tfVariables map[string]tfVariable

func GenerateVariables(ctx *cli.Context) (string, string, error) {
	variableBlocks := make(tfVariables)
	usedVars := make(map[string]struct{})

	entries, err := os.ReadDir(ctx.String("dir"))
	if err != nil {
		return "", "", fmt.Errorf("path %v: %w", ctx.String("dir"), err)
	}
	for _, entry := range entries {
		path := filepath.Join(ctx.String("dir"), entry.Name())
		if entry.IsDir() || !strings.HasSuffix(path, ".tf") {
			return "", "", nil
		}
		file, err := os.Open(path)
		if err != nil {
			return "", "", fmt.Errorf("open %v: %w", path, err)
		}
		defer file.Close()

		var copiedFile bytes.Buffer
		teeFile := io.TeeReader(file, &copiedFile)

		declaredVars, err := collectDeclaredVariables(teeFile)
		if err != nil {
			return "", "", fmt.Errorf("collect declared variables: %w", err)
		}
		for k, v := range declaredVars {
			variableBlocks[k] = v
		}

		vu, err := collectUsedVariables(&copiedFile)
		if err != nil {
			return "", "", fmt.Errorf("collect used variables: %w", err)
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
			}
		} else {
			variableBlocks[used] = tfVariable{
				block: fmt.Sprintf(outputTemplate, used),
				used:  true,
			}
		}
	}

	if ctx.Bool("sync") {
		var keysToDelete []string
		for k, v := range variableBlocks {
			if !v.used {
				keysToDelete = append(keysToDelete, k)
			}
		}
		for _, key := range keysToDelete {
			delete(variableBlocks, key)
		}
	}

	var output []string
	for _, v := range variableBlocks {
		output = append(output, v.block)
	}

	return strings.Join(output, "\n\n"), "", nil
}

func collectDeclaredTfvars(path string) (map[string]struct{}, error) {
	var config map[string]*hcl.Attribute
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, fmt.Errorf("decode file: %w", err)
	}
	for k, v := range config {
		fmt.Printf("Configuration is %v: %v %T\n", k, v.Expr, v)
	}
	return nil, nil
}
