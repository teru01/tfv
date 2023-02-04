package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

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
	var variableBlocks tfVariables
	var usedVars map[string]struct{}

	entries, err := os.ReadDir(ctx.String("dir"))
	if err != nil {
		return "", "", fmt.Errorf("path %v: %w", ctx.String("dir"), err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			return "", "", nil
		}
		file, err := os.Open(entry.Name())
		if err != nil {
			return "", "", fmt.Errorf("open %v: %w", entry.Name(), err)
		}
		defer file.Close()

		var copiedFile *bytes.Buffer
		teeFile := io.TeeReader(file, copiedFile)

		declaredVars, err := collectDeclaredVariables(teeFile)
		if err != nil {
			return "", "", fmt.Errorf("collect declared variables: %w", err)
		}
		for k, v := range declaredVars {
			variableBlocks[k] = v
		}

		vu, err := collectUsedVariables(copiedFile)
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

	var b strings.Builder
	for _, v := range variableBlocks {
		fmt.Fprintf(&b, v.block+"\n")
	}

	return b.String(), "", nil
}
