package main

import "path/filepath"

func readRust(root string) *RustInfo {
	f := filepath.Join(root, "Cargo.toml")
	if !exists(f) {
		return nil
	}
	j := parseTomlLight(f, 64*1024)
	ri := &RustInfo{
		Package: toStr(nested(j, "package", "name")),
		Edition: toStr(nested(j, "package", "edition")),
		Deps:    toStrMap(nested(j, "dependencies")),
	}
	if nested(j, "workspace") != nil {
		ri.Workspace = true
	}
	return ri
}
