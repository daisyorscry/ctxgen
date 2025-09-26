package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func parseTomlLight(full string, max int) map[string]any {
	out := map[string]any{}
	b, err := os.ReadFile(full)
	if err != nil {
		return out
	}
	if len(b) > max {
		b = b[:max]
	}
	section := ""
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(strings.Trim(line, "[]"))
			if _, ok := out[section]; !ok {
				out[section] = map[string]any{}
			}
			continue
		}
		if i := strings.IndexByte(line, '='); i > 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			v = strings.Trim(v, `"'`)
			if section == "" {
				out[k] = v
			} else {
				m := out[section].(map[string]any)
				m[k] = v
			}
		}
	}
	return out
}

func parseYAMLLight(full string, max int) map[string]any {
	out := map[string]any{}
	b, err := os.ReadFile(full)
	if err != nil {
		return out
	}
	if len(b) > max {
		b = b[:max]
	}
	type stackFrame struct {
		indent int
		m      map[string]any
	}
	root := map[string]any{}
	stack := []stackFrame{{indent: -1, m: root}}
	lines := bytes.Split(b, []byte("\n"))
	for _, raw := range lines {
		line := string(raw)
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		indent := 0
		for i := 0; i < len(line) && line[i] == ' '; i++ {
			indent++
		}
		line = strings.TrimSpace(line)
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, `"'`)
		for len(stack) > 0 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}
		cur := stack[len(stack)-1].m
		if v == "" {
			child := map[string]any{}
			cur[k] = child
			stack = append(stack, stackFrame{indent: indent, m: child})
		} else {
			cur[k] = v
		}
	}
	return root
}

func writeNDJSON(root string, m *Manifest, outPath string, csvExt string, withSHA1 bool) error {
	allow := make(map[string]bool)
	for _, e := range strings.Split(csvExt, ",") {
		e = strings.ToLower(strings.TrimSpace(e))
		e = strings.TrimPrefix(e, ".")
		if e == "" {
			continue
		}
		allow[e] = true
	}

	type F struct {
		Path string
		Size int64
		Lang string
	}
	var files []F
	total := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
		info, e := d.Info()
		if e != nil {
			return nil
		}
		total++

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(rel), "."))
		if !allow[ext] {
			return nil
		}

		files = append(files, F{Path: rel, Size: info.Size(), Lang: ext})
		return nil
	})
	if err != nil {
		return err
	}

	slices.SortFunc(files, func(a, b F) int { return strings.Compare(a.Path, b.Path) })

	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	fh, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer fh.Close()

	gw := gzip.NewWriter(fh)
	defer gw.Close()

	bw := bufio.NewWriter(gw)
	defer bw.Flush()

	enc := json.NewEncoder(bw)

	toc := NDJSONTOC{
		Type:        "toc",
		Project:     m.Project,
		Root:        m.Root,
		GeneratedAt: m.GeneratedAt,
		FilesTotal:  total,
		Files:       make([]FileEntry, 0, len(files)),
		ManifestLite: struct {
			Framework     string          `json:"framework,omitempty"`
			Composer      *ComposerInfo   `json:"composer,omitempty"`
			Node          *NodeInfo       `json:"node,omitempty"`
			Go            *GoInfo         `json:"go,omitempty"`
			EnvKeys       []string        `json:"env_keys,omitempty"`
			CodeSummary   *CodeSummary    `json:"code_summary,omitempty"`
			Laravel       *LaravelCtx     `json:"laravel,omitempty"`
			Routes        []RouteFile     `json:"routes,omitempty"`
			Migrations    []string        `json:"migrations,omitempty"`
			Seeders       []string        `json:"seeders,omitempty"`
			Git           *GitInfo        `json:"git,omitempty"`
			CustomSignals map[string]bool `json:"custom_signals,omitempty"`
		}{
			Framework:     m.Framework,
			Composer:      m.Composer,
			Node:          m.Node,
			Go:            m.Go,
			EnvKeys:       m.EnvKeys,
			CodeSummary:   m.CodeSummary,
			Laravel:       m.Laravel,
			Routes:        m.Routes,
			Migrations:    m.Migrations,
			Seeders:       m.Seeders,
			Git:           m.Git,
			CustomSignals: m.CustomSignals,
		},
	}
	for _, f := range files {
		entry := FileEntry{Path: f.Path, Size: f.Size, Lang: f.Lang}
		if withSHA1 {
			if b, err := os.ReadFile(filepath.Join(root, f.Path)); err == nil {
				h := sha1.Sum(b)
				entry.SHA1 = hex.EncodeToString(h[:])
			}
		}
		toc.Files = append(toc.Files, entry)
	}
	if err := enc.Encode(&toc); err != nil {
		return err
	}

	buf := make([]byte, 0, 64*1024)
	for _, f := range files {
		full := filepath.Join(root, f.Path)
		sf, err := os.Open(full)
		if err != nil {
			continue
		}
		b, err := io.ReadAll(sf)
		_ = sf.Close()
		if err != nil {
			continue
		}
		rec := NDJSONFile{
			Type:    "file",
			Path:    f.Path,
			Size:    int64(len(b)),
			Lang:    f.Lang,
			Content: string(b),
		}
		if withSHA1 {
			h := sha1.Sum(b)
			rec.SHA1 = hex.EncodeToString(h[:])
		}
		if err := enc.Encode(&rec); err != nil {
			return err
		}
		_ = buf
	}
	return nil
}

func outJSON(out string, v any) {
	data, _ := json.MarshalIndent(v, "", "  ")
	if out == "" {
		os.Stdout.Write(data)
		return
	}
	_ = os.MkdirAll(filepath.Dir(out), 0o755)
	_ = os.WriteFile(out, data, 0o644)
}
