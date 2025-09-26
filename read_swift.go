package main

import (
	"os"
	"path/filepath"
	"regexp"
)

func readSwift(root string) *SwiftInfo {
	if exists(filepath.Join(root, "Package.swift")) {
		b, _ := os.ReadFile(filepath.Join(root, "Package.swift"))
		si := &SwiftInfo{Deps: map[string]string{}}
		reName := regexp.MustCompile(`name:\s*"([^"]+)"`)
		if m := reName.FindSubmatch(b); len(m) == 2 {
			si.PackageName = string(m[1])
		}
		reDep := regexp.MustCompile(`\.package\(.*?url:\s*"([^"]+)"(?:,\s*from:\s*"([^"]+)")?.*?\)`)
		for _, mm := range reDep.FindAllSubmatch(b, -1) {
			url := string(mm[1])
			ver := ""
			if len(mm) == 3 {
				ver = string(mm[2])
			}
			si.Deps[url] = ver
		}
		return si
	}
	if exists(filepath.Join(root, "Podfile")) {
		if _, err := os.Stat(filepath.Join(root, "Package.swift")); err != nil {
			return &SwiftInfo{UsesCocoaPods: true, Deps: map[string]string{}}
		}
	}
	return nil
}
