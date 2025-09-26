package main

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func scanProject(root string, routeMaxLines int) (*CodeSummary, *LaravelCtx, []RouteFile, []string, []string) {
	sum := &CodeSummary{Langs: map[string]int{}}
	lctx := &LaravelCtx{}
	var routes []RouteFile
	var migrations []string
	var seeders []string

	reNS := regexp.MustCompile(`(?m)^namespace\s+([^;]+);`)
	reClass := regexp.MustCompile(`(?m)^(?:abstract\s+|final\s+)?class\s+([A-Za-z0-9_\\]+)`)
	reMeth := regexp.MustCompile(`(?m)^(\s*)public\s+function\s+([A-Za-z0-9_]+)\s*\(`)

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

		ext := strings.ToLower(filepath.Ext(rel))
		switch ext {
		case ".php", ".js", ".ts", ".tsx", ".jsx", ".go", ".json", ".env", ".sql", ".rb", ".py", ".rs", ".java", ".kt", ".cs", ".dart", ".swift", ".m", ".mm", ".yaml", ".yml", ".xml", ".gradle", ".kts", ".md", ".txt", ".sh":
			sum.Files++
			if ext != "" {
				sum.Langs[strings.TrimPrefix(ext, ".")]++
			}
		default:
			return nil
		}

		if ext == ".php" {
			sum.PHPLOC += countLOC(path)
		}

		if strings.HasPrefix(rel, "app/") && ext == ".php" {
			switch {
			case strings.HasPrefix(rel, "app/Http/Controllers/"):
				pc := parsePHP(path, reNS, reClass, reMeth)
				pc.Path = rel
				lctx.Controllers = append(lctx.Controllers, pc)
			case strings.HasPrefix(rel, "app/Http/Middleware/"):
				pc := parsePHP(path, reNS, reClass, reMeth)
				pc.Path = rel
				lctx.Middleware = append(lctx.Middleware, pc)
			case strings.HasPrefix(rel, "app/Models/"):
				pc := parsePHP(path, reNS, reClass, reMeth)
				pc.Path = rel
				lctx.Models = append(lctx.Models, pc)
			case strings.HasPrefix(rel, "app/Traits/"):
				pc := parsePHP(path, reNS, reClass, reMeth)
				pc.Path = rel
				lctx.Traits = append(lctx.Traits, pc)
			case strings.HasPrefix(rel, "app/Helpers/"):
				pc := parsePHP(path, reNS, reClass, reMeth)
				pc.Path = rel
				lctx.Helpers = append(lctx.Helpers, pc)
			}
		}

		if strings.HasPrefix(rel, "routes/") && (ext == ".php" || ext == ".ts" || ext == ".js") {
			routes = append(routes, readRouteFile(path, rel, routeMaxLines))
		}
		if strings.HasPrefix(rel, "database/migrations/") && ext == ".php" {
			migrations = append(migrations, rel)
		}
		if strings.HasPrefix(rel, "database/seeders/") && ext == ".php" {
			seeders = append(seeders, rel)
		}
		return nil
	})

	slices.SortFunc(lctx.Controllers, func(a, b PHPClassFile) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(lctx.Middleware, func(a, b PHPClassFile) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(lctx.Models, func(a, b PHPClassFile) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(lctx.Traits, func(a, b PHPClassFile) int { return strings.Compare(a.Path, b.Path) })
	slices.SortFunc(lctx.Helpers, func(a, b PHPClassFile) int { return strings.Compare(a.Path, b.Path) })

	slices.SortFunc(routes, func(a, b RouteFile) int { return strings.Compare(a.Path, b.Path) })
	slices.Sort(migrations)
	slices.Sort(seeders)
	return sum, lctx, routes, migrations, seeders
}
