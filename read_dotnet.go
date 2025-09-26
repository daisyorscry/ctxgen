package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func readDotNet(root string) *DotNetInfo {
	var projs []DotNetProject
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if strings.HasSuffix(rel, ".csproj") {
			pp := DotNetProject{Path: rel, Refs: map[string]string{}}
			b, _ := os.ReadFile(path)
			reSDK := regexp.MustCompile(`Sdk="([^"]+)"`)
			if m := reSDK.FindSubmatch(b); len(m) == 2 {
				pp.SDK = string(m[1])
			}
			rePkg := regexp.MustCompile(`(?s)<PackageReference\s+Include="([^"]+)"\s+Version="([^"]+)"`)
			for _, mm := range rePkg.FindAllSubmatch(b, -1) {
				pp.Refs[string(mm[1])] = string(mm[2])
			}
			projs = append(projs, pp)
		}
		return nil
	})
	if len(projs) == 0 {
		return nil
	}
	return &DotNetInfo{Projects: projs}
}
