package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pattern = regexp.MustCompile(`var\.([^}")\[\],]*)`)

func main() {
	vars := make(map[string]struct{})
	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".tf") {
			return nil
		}
		varsInFile, err := extractVars(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range varsInFile {
			vars[v] = struct{}{}
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	for v := range vars {
		fmt.Printf(`variable "%s" {
  description = ""
}`+"\n\n", v)
	}
}

func extractVars(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var vars []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			vars = append(vars, match[1])
		}
	}
	return vars, nil
}
