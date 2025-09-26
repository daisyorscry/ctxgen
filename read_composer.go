package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func readComposer(root string) *ComposerInfo {
	f := filepath.Join(root, "composer.json")
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var raw rawJSON
	if json.Unmarshal(b, &raw) != nil {
		return nil
	}
	return &ComposerInfo{
		Name:         toStr(raw["name"]),
		Type:         toStr(raw["type"]),
		Require:      toStrMap(raw["require"]),
		RequireDev:   toStrMap(raw["require-dev"]),
		AutoloadPSR4: toStrMap(nested(raw, "autoload", "psr-4")),
	}
}

func phpHas(l *LaravelCtx, sub string) bool {
	if l == nil {
		return false
	}
	sub = strings.ToLower(sub)
	for _, set := range [][]PHPClassFile{l.Controllers, l.Middleware, l.Models, l.Traits, l.Helpers} {
		for _, x := range set {
			if strings.Contains(strings.ToLower(x.Class), sub) || strings.Contains(strings.ToLower(x.Path), sub) {
				return true
			}
		}
	}
	return false
}
func hasComposer(c *ComposerInfo, name string) bool {
	if c == nil {
		return false
	}
	if _, ok := c.Require[name]; ok {
		return true
	}
	if _, ok := c.RequireDev[name]; ok {
		return true
	}
	return false
}

func parsePHP(full string, reNS, reClass, reMeth *regexp.Regexp) PHPClassFile {
	b, _ := os.ReadFile(full)
	ns := firstGroup(reNS.FindSubmatch(b), 1)
	cls := firstGroup(reClass.FindSubmatch(b), 1)
	var methods []string
	for _, m := range reMeth.FindAllSubmatch(b, -1) {
		methods = append(methods, string(m[2]))
	}
	return PHPClassFile{Namespace: ns, Class: cls, Methods: methods}
}

type rawJSON = map[string]any

func looksLaravel(comp *ComposerInfo, root string) bool {
	if comp != nil {
		if _, ok := comp.Require["laravel/framework"]; ok {
			return true
		}
	}
	if _, err := os.Stat(filepath.Join(root, "artisan")); err == nil {
		return true
	}
	return false
}
