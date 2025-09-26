package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var routeScanExt = map[string]bool{
	".php": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
	".go": true, ".py": true, ".rb": true, ".java": true, ".cs": true, ".kt": true, ".rs": true,
}

var lightRouteIndicators = []string{
	"Route::", "->middleware(", "->name(",
	".get(", ".post(", ".put(", ".patch(", ".delete(", "@Get(", "@Post(", "@Put(", "@Patch(", "@Delete(",
	"@app.get(", "@app.post(", "@app.put(", "@app.patch(", "@app.delete(", "path(", "re_path(",
	".GET(", ".POST(", ".PUT(", ".PATCH(", ".DELETE(", "HandleFunc(", ".Methods(",
	"@GetMapping(", "@PostMapping(", "@PutMapping(", "@PatchMapping(", "@DeleteMapping(", "@RequestMapping(",
	"[HttpGet", "[HttpPost", "[HttpPut", "[HttpPatch", "[HttpDelete", "[Route(",
	" get '", " post '", " put '", " patch '", " delete '", "resources ",
}

func looksRouteText(path string) bool {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "/routes/") ||
		strings.Contains(lower, "/route") ||
		strings.HasSuffix(lower, "routes.php") ||
		strings.HasSuffix(lower, "routes.rb") ||
		strings.HasSuffix(lower, "urls.py") ||
		strings.Contains(lower, "/controller") ||
		strings.Contains(lower, "controller.") ||
		strings.Contains(lower, "/handlers/") ||
		strings.Contains(lower, "/api/") {
		return true
	}
	return false
}

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

		// exclude berat
		if d.IsDir() {
			if hasAnyPrefix(rel+"/",
				".git/", ".idea/", ".vscode/", "node_modules/", "vendor/", "storage/", "bootstrap/cache/", "public/", ".next/", "dist/", "build/",
			) {
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

		// PHP LOC
		if ext == ".php" {
			sum.PHPLOC += countLOC(path)
		}

		// ================= Laravel deep context (opsional, jika ada) =================
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

		// ================= General routes discovery (multi-framework) =================
		if routeScanExt[ext] {
			// fast path: file “kemungkinan” berisi route berdasarkan nama, atau
			// kalau di folder routes, urls.py, controllers, api, dsb → parse
			if looksRouteText(rel) {
				routes = append(routes, readRouteFile(path, rel, routeMaxLines))
			} else {
				// heuristik super ringan: cek beberapa byte pertama untuk indikator
				// (menghindari baca seluruh file besar di sini—parsing penuh di readRouteFile)
				// trade-off: tetap efisien tapi cukup sensitif.
				peekN := 32 * 1024
				b, _ := osReadHead(path, peekN)
				text := string(b)
				for _, key := range lightRouteIndicators {
					if strings.Contains(text, key) {
						routes = append(routes, readRouteFile(path, rel, routeMaxLines))
						break
					}
				}
			}
		}

		// ================= Migrations (multi-ecosystem) =================
		// Laravel
		if strings.HasPrefix(rel, "database/migrations/") && ext == ".php" {
			migrations = append(migrations, rel)
		}
		// Rails
		if strings.HasPrefix(rel, "db/migrate/") && ext == ".rb" {
			migrations = append(migrations, rel)
		}
		// Django (skip __init__.py)
		if strings.Contains(rel, "/migrations/") && ext == ".py" && !strings.HasSuffix(rel, "__init__.py") {
			migrations = append(migrations, rel)
		}
		// Node (Sequelize/Knex)
		if strings.HasPrefix(rel, "migrations/") && (ext == ".js" || ext == ".ts") {
			migrations = append(migrations, rel)
		}
		// Prisma: direktori “prisma/migrations/**” (filenya bisa berbagai ekstensi)
		if strings.HasPrefix(rel, "prisma/migrations/") {
			migrations = append(migrations, rel)
		}

		// ================= Seeders (multi-ecosystem) =================
		// Laravel
		if strings.HasPrefix(rel, "database/seeders/") && ext == ".php" {
			seeders = append(seeders, rel)
		}
		// Rails
		if rel == "db/seeds.rb" {
			seeders = append(seeders, rel)
		}
		// Node (Sequelize)
		if strings.HasPrefix(rel, "seeders/") && (ext == ".js" || ext == ".ts") {
			seeders = append(seeders, rel)
		}
		// Django fixtures (opsional): **/fixtures/*.json
		if strings.Contains(rel, "/fixtures/") && strings.HasSuffix(rel, ".json") {
			seeders = append(seeders, rel)
		}

		return nil
	})

	// sorting
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

// baca head file terbatas (untuk heuristik indikator)
func osReadHead(full string, n int) ([]byte, error) {
	f, err := os.Open(full)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, n)
	k, _ := f.Read(buf)
	return buf[:k], nil
}
