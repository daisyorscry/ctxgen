package main

import (
	"os"
	"path/filepath"
	"regexp"
)

func readRuby(root string) *RubyInfo {
	f := filepath.Join(root, "Gemfile.lock")
	if !exists(f) {
		return nil
	}
	b, _ := os.ReadFile(f)
	re := regexp.MustCompile(`\s{4}([A-Za-z0-9_\-]+)\s\(([^)]+)\)`)
	gems := map[string]string{}
	for _, mm := range re.FindAllSubmatch(b, -1) {
		gems[string(mm[1])] = string(mm[2])
	}
	if len(gems) == 0 {
		return nil
	}
	return &RubyInfo{Gems: gems}
}
