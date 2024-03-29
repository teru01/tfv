package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

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
	return generateVariables(ctx.Bool("sync"), ctx.String("dir"), ctx.String("variables-file"), ctx.String("tfvars-file"))
}

func generateVariables(sync bool, dir, variablesFile, tfVarsFile string) (string, string, error) {
	usedVars, err := collectAllUsedVariables(dir)
	if err != nil {
		return "", "", fmt.Errorf("collect used vars: %w", err)
	}

	variables, keysToDelete, err := rebuildVariableFile(usedVars, variablesFile, sync)
	if err != nil {
		return "", "", fmt.Errorf("build variables blocks: %w", err)
	}

	filename := tfVarsFile
	if filename != "" && sync {
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
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("open %v: %w", path, err)
	} else {
		defer file.Close()
	}

	if errors.Is(err, os.ErrNotExist) {
		l := generateUndeclaredVariables(usedVars)
		return strings.Join(l, "\n"), make(map[string]struct{}), nil
	}
	declaredVars, unusedVariables, err := rebuildDeclaredVariables(file, usedVars, sync)
	if err != nil {
		return "", nil, fmt.Errorf("collect declared variables: %w", err)
	}

	return declaredVars, unusedVariables, nil
}
