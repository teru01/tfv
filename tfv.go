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

func Execute(c *cli.Context) error {
	var declaredVars tfVariables
	var usedVars map[string]struct{}

	entries, err := os.ReadDir(c.String("dir"))
	if err != nil {
		return fmt.Errorf("path %v: %w", c.String("dir"), err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			return nil
		}
		file, err := os.Open(entry.Name())
		if err != nil {
			return fmt.Errorf("open %v: %w", entry.Name(), err)
		}
		defer file.Close()

		var copiedFile *bytes.Buffer
		teeFile := io.TeeReader(file, copiedFile)

		vd, err := collectDeclaredVariables(teeFile)
		if err != nil {
			return fmt.Errorf("collect declared variables: %w", err)
		}
		for k, v := range vd {
			declaredVars[k] = v
		}

		vu, err := collectUsedVariables(copiedFile)
		if err != nil {
			return fmt.Errorf("collect used variables: %w", err)
		}
		for k := range vu {
			usedVars[k] = struct{}{}
		}
	}

	for used := range usedVars {
		if _, ok := declaredVars[used]; ok {

		}
	}

	return nil
}
