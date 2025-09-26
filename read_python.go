package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func readPython(root string) *PythonInfo {
	out := &PythonInfo{}
	if exists(filepath.Join(root, "requirements.txt")) {
		out.HasPip = true
		out.Requirements = parseRequirements(filepath.Join(root, "requirements.txt"))
	}
	if exists(filepath.Join(root, "pyproject.toml")) {
		out.HasPoetry = true
		out.PyProject = parseTomlLight(filepath.Join(root, "pyproject.toml"), 64*1024)
	}
	if exists(filepath.Join(root, "Pipfile")) {
		out.HasPipenv = true
	}
	if !out.HasPip && !out.HasPoetry && !out.HasPipenv {
		return nil
	}
	return out
}

func parseRequirements(full string) map[string]string {
	b, err := os.ReadFile(full)
	if err != nil {
		return nil
	}
	m := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(b))
	re := regexp.MustCompile(`^\s*([A-Za-z0-9_\-\.]+)\s*(?:==|>=|<=|~=|!=)\s*([A-Za-z0-9_\-\.]+)`)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pkg := line
		ver := ""
		if mm := re.FindStringSubmatch(line); len(mm) == 3 {
			pkg = mm[1]
			ver = mm[2]
		}
		if pkg != "" {
			m[pkg] = ver
		}
	}
	return m
}
