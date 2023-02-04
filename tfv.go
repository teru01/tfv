package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
)

var outputTemplate = `variable "%s" {
	description = ""
  }`

var pattern = regexp.MustCompile(`var\.([^}")\[\],\s]*)`)
var pattrnForExtractVariables = regexp.MustCompile(`variable\s*"(.*)"`)

func Execute(c *cli.Context) error {
	err := filepath.WalkDir(c.String("dir"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("path %v: %w", path, err)
		}

		if !strings.HasSuffix(path, ".tf") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %v: %w", path, err)
		}
		defer file.Close()

		var copiedFile *bytes.Buffer
		teeFile := io.TeeReader(file, copiedFile)

		declaredVars, err := collectDeclaredVariables(teeFile)
		if err != nil {
			return fmt.Errorf("collect declared variables: %w", err)
		}
		usedVars, err := collectUsedVariables(copiedFile)
		if err != nil {
			return fmt.Errorf("collect used variables: %w", err)
		}

		fmt.Println(declaredVars, usedVars)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk files: %w", err)
	}
	return nil
}

// カレントディレクトリのtfファイルからvar.fooを抽出し、変数名を返す
func walkFiles() ([]string, error) {
	vars := make(map[string]struct{})
	varsDeclared := make([]string, 0)
	entries, err := os.ReadDir("./")
	if err != nil {
		return nil, fmt.Errorf("failed to readDir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		file, err := os.Open(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		buf := &bytes.Buffer{}
		tee := io.TeeReader(file, buf)

		varsInFile, err := extractVars(tee)
		if err != nil {
			return nil, fmt.Errorf("failed to extract vars: %w", err)
		}
		for _, v := range varsInFile {
			vars[v] = struct{}{}
		}
		vd, err := extractAlreadyDeclared(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to extract already declared vars: %w", err)
		}
		varsDeclared = append(varsDeclared, vd...)
		file.Close()
	}
	results := make([]string, 0)
	i := 0
	for v := range vars {
		var declared bool
		for _, vd := range varsDeclared {
			if v == vd {
				declared = true
				break
			}
		}
		if !declared {
			results = append(results, v)
			i++
		}
	}
	return results, nil
}

// extract foo from `variables "foo"` and return it
func extractAlreadyDeclared(file io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(file)
	var vars []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "variable") {
			s := extractVariable(line)
			if s != "" {
				vars = append(vars, s)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan: %w", err)
	}
	return vars, nil
}

// impl extractVariable
func extractVariable(line string) string {
	matches := pattrnForExtractVariables.FindStringSubmatch(line)
	if len(matches) == 0 {
		return ""
	}
	return strings.TrimSpace(matches[1])
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
