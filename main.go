package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

var outputTemplate = `variable "%s" {
  description = ""
}`

var pattern = regexp.MustCompile(`var\.([^}")\[\],]*)`)

func main() {
	app := &cli.App{
		Name:   "tfv",
		Usage:  "generate Terraform variables declaration",
		Action: printVariables,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func printVariables(c *cli.Context) error {
	vars, err := walkFiles()
	if err != nil {
		return xerrors.Errorf("failed to walk files: %w", err)
	}
	for _, v := range vars {
		fmt.Printf(outputTemplate+"\n\n", v)
	}
	return nil
}

// カレントディレクトリのtfファイルからvar.fooを抽出し、変数名を返す
func walkFiles() ([]string, error) {
	vars := make(map[string]struct{})
	entries, err := os.ReadDir("./")
	if err != nil {
		return nil, xerrors.Errorf("failed to readDir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		file, err := os.Open(entry.Name())
		if err != nil {
			return nil, xerrors.Errorf("failed to open file: %w", err)
		}
		varsInFile, err := extractVars(file)
		if err != nil {
			return nil, xerrors.Errorf("failed to extract vars: %w", err)
		}
		for _, v := range varsInFile {
			vars[v] = struct{}{}
		}
		file.Close()
	}
	results := make([]string, len(vars))
	i := 0
	for v := range vars {
		results[i] = v
		i++
	}
	return results, nil
}

func extractVars(file io.Reader) ([]string, error) {
	var vars []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			vars = append(vars, match[1])
		}
	}
	return vars, nil
}
