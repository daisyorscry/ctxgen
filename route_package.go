package main

import (
	"bufio"
	"bytes"
	"os"
	"regexp"
	"strings"
)

func readRouteFile(full, rel string, routeMaxLines int) RouteFile {
	out := RouteFile{Path: rel}
	b, _ := os.ReadFile(full)
	sc := bufio.NewScanner(bytes.NewReader(b))
	count := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "Route::") || strings.Contains(line, "->middleware(") || strings.Contains(line, "->name(") {
			out.Snips = append(out.Snips, truncate(line, 240))
		}
		count++
		if count >= routeMaxLines {
			break
		}
	}
	re := regexp.MustCompile(`Route::(get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		out.Guessed = append(out.Guessed, strings.ToUpper(m[1])+" "+m[2])
	}
	return out
}
