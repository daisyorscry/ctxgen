package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func toStr(x any) string {
	if s, ok := x.(string); ok {
		return s
	}
	return ""
}
func toStrMap(x any) map[string]string {
	out := map[string]string{}
	m, ok := x.(map[string]any)
	if !ok {
		return out
	}
	for k, v := range m {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func nested(m map[string]any, path ...string) any {
	cur := any(m)
	for _, p := range path {
		m2, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m2[p]
	}
	return cur
}
func countLOC(full string) int {
	f, err := os.Open(full)
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	n := 0
	for sc.Scan() {
		n++
	}
	return n
}

func hasAnyPrefix(s string, prefs ...string) bool {
	for _, p := range prefs {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
func unique[T comparable](in []T) []T {
	seen := map[T]bool{}
	out := make([]T, 0, len(in))
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
func firstExist(root string, names []string) string {
	for _, n := range names {
		if exists(filepath.Join(root, n)) {
			return n
		}
	}
	return ""
}
func filesExist(root string, globs []string) bool {
	found := false
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, g := range globs {
			ok, _ := filepath.Match(g, rel)
			if ok {
				found = true
				break
			}
		}
		return nil
	})
	return found
}

func firstGroup(m [][]byte, idx int) string {
	if len(m) == 0 {
		return ""
	}
	return string(m[idx])
}

func hasAnyKey(m map[string]string, key string) bool { _, ok := m[key]; return ok }

func embedSamples(root, csvGlobs string, maxBytes int, maxFiles int) []SampleFile {
	var patterns []string
	for _, g := range strings.Split(csvGlobs, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			patterns = append(patterns, g)
		}
	}
	if len(patterns) == 0 {
		return nil
	}

	var matches []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, pat := range patterns {
			ok, _ := filepath.Match(pat, rel)
			if ok {
				matches = append(matches, rel)
				break
			}
		}
		return nil
	})
	slices.Sort(matches)
	if len(matches) > maxFiles {
		matches = matches[:maxFiles]
	}

	var out []SampleFile
	for _, rel := range matches {
		full := filepath.Join(root, rel)
		b, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		sf := SampleFile{Path: rel, Bytes: len(b)}
		if len(b) > maxBytes {
			sf.Truncated = true
			sf.Content = string(b[:maxBytes])
		} else {
			sf.Content = string(b)
		}
		out = append(out, sf)
	}
	return out
}
