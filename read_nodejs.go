package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func readPackageJSON(root string) *NodeInfo {
	f := filepath.Join(root, "package.json")
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var raw rawJSON
	if json.Unmarshal(b, &raw) != nil {
		return nil
	}

	ni := &NodeInfo{
		Name:            toStr(raw["name"]),
		Scripts:         toStrMap(raw["scripts"]),
		Dependencies:    toStrMap(raw["dependencies"]),
		DevDependencies: toStrMap(raw["devDependencies"]),
	}
	ni.Next = hasAnyKey(ni.Dependencies, "next") || hasAnyKey(ni.DevDependencies, "next")
	ni.Expo = hasAnyKey(ni.Dependencies, "expo")
	ni.ReactNative = hasAnyKey(ni.Dependencies, "react-native")
	ni.Typescript = hasAnyKey(ni.DevDependencies, "typescript") || hasAnyKey(ni.Dependencies, "typescript")
	if emptyNode(ni) {
		return nil
	}
	return ni
}

func nodeDepsHas(n *NodeInfo, name string) bool {
	if n == nil {
		return false
	}
	if _, ok := n.Dependencies[name]; ok {
		return true
	}
	if _, ok := n.DevDependencies[name]; ok {
		return true
	}
	return false
}
func nodeHasScript(n *NodeInfo, script string) bool {
	if n == nil {
		return false
	}
	_, ok := n.Scripts[script]
	return ok
}

func pkgHas(n *NodeInfo, keys []string) bool {
	if n == nil {
		return false
	}
	for _, k := range keys {
		if _, ok := n.Dependencies[k]; ok {
			return true
		}
		if _, ok := n.DevDependencies[k]; ok {
			return true
		}
	}
	return false
}

func emptyNode(n *NodeInfo) bool {
	return n.Name == "" && len(n.Scripts) == 0 && len(n.Dependencies) == 0 && len(n.DevDependencies) == 0
}

func buildFilesTOC(root string, maxFiles int, withSHA1 bool) (int, []FileEntry) {
	var total int
	var out []FileEntry
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if hasAnyPrefix(rel+"/", ".git/", ".idea/", ".vscode/", "node_modules/", "vendor/", "storage/", "bootstrap/cache/", "public/", ".next/", "dist/", "build/") {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		total++

		ext := strings.ToLower(filepath.Ext(rel))
		lang := strings.TrimPrefix(ext, ".")
		entry := FileEntry{Path: rel, Size: info.Size(), Lang: lang}

		if withSHA1 {
			if b, err := os.ReadFile(path); err == nil {
				h := sha1.Sum(b)
				entry.SHA1 = hex.EncodeToString(h[:])
			}
		}

		if len(out) < maxFiles {
			out = append(out, entry)
		}
		return nil
	})

	slices.SortFunc(out, func(a, b FileEntry) int { return strings.Compare(a.Path, b.Path) })
	return total, out
}
