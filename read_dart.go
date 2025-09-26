package main

import "path/filepath"

func readDart(root string) *DartInfo {
	f := filepath.Join(root, "pubspec.yaml")
	if !exists(f) {
		return nil
	}
	j := parseYAMLLight(f, 128*1024)
	di := &DartInfo{
		Name:         toStr(nested(j, "name")),
		Dependencies: toStrMap(nested(j, "dependencies")),
		DevDeps:      toStrMap(nested(j, "dev_dependencies")),
	}
	if nested(j, "flutter") != nil {
		di.Flutter = true
	}
	if di.Name == "" && len(di.Dependencies) == 0 && len(di.DevDeps) == 0 && !di.Flutter {
		return nil
	}
	return di
}
