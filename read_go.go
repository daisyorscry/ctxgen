package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func readGoModule(root string) *GoInfo {
	f := filepath.Join(root, "go.mod")
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	mod := ""
	var req []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	inReqBlock := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "module ") {
			mod = strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
		if strings.HasPrefix(line, "require (") {
			inReqBlock = true
			continue
		}
		if inReqBlock && strings.HasPrefix(line, ")") {
			inReqBlock = false
			continue
		}
		if inReqBlock {
			fields := strings.Fields(line)
			if len(fields) >= 1 {
				req = append(req, fields[0])
			}
		}
		if strings.HasPrefix(line, "require ") && !strings.HasPrefix(line, "require (") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 1 {
				req = append(req, parts[0])
			}
		}
	}
	if mod == "" && len(req) == 0 {
		return nil
	}
	return &GoInfo{Module: mod, Requires: unique(req)}
}

func goHas(g *GoInfo, pkgs []string) bool {
	if g == nil {
		return false
	}
	for _, r := range g.Requires {
		for _, k := range pkgs {
			if strings.HasPrefix(r, k) {
				return true
			}
		}
	}
	return false
}
