package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func listEnvKeys(root string) []string {
	f := filepath.Join(root, ".env")
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var keys []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '='); i > 0 {
			k := strings.TrimSpace(line[:i])
			if k != "" {
				keys = append(keys, k)
			}
		}
	}
	slices.Sort(keys)
	return unique(keys)
}
