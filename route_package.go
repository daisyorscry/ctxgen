package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func joinPath(prefix, p string) string {
	if prefix == "" {
		if p == "" {
			return "/"
		}
		if !strings.HasPrefix(p, "/") {
			return "/" + p
		}
		return p
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if p == "" || p == "/" {
		return prefix
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	out := prefix + p
	out = strings.ReplaceAll(out, "//", "/")
	return out
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func addGuess(set map[string]struct{}, out *[]string, method, path string) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "ANY"
	}
	if path == "" {
		path = "/"
	}
	key := method + " " + path
	if _, ok := set[key]; !ok {
		set[key] = struct{}{}
		*out = append(*out, key)
	}
}

func stripInlineComments(line string) string {
	l := strings.TrimSpace(line)
	if i := strings.Index(l, "//"); i >= 0 {
		l = l[:i]
	}
	if i := strings.Index(l, "#"); i >= 0 {
		l = l[:i]
	}
	if i := strings.Index(l, "//"); i >= 0 {
		l = l[:i]
	}
	return strings.TrimSpace(l)
}

var (
	// Laravel
	reLaravelRoute  = regexp.MustCompile(`Route::(get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)
	reLaravelGroup1 = regexp.MustCompile(`Route::prefix\(\s*['"]([^'"]+)['"]\s*\)\s*->group`)
	reLaravelGroup2 = regexp.MustCompile(`Route::group\(\s*\[\s*'prefix'\s*=>\s*['"]([^'"]+)['"]\s*\]`)

	// Express / Koa
	reExpressRoute = regexp.MustCompile(`\b(?:app|router|\w+)\.(get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)
	reExpressUse   = regexp.MustCompile(`\bapp\.use\s*\(\s*['"]([^'"]+)['"]\s*,\s*(\w+)\s*\)`)

	// NestJS
	reNestController = regexp.MustCompile(`@Controller\(\s*['"]([^'"]*)['"]?\s*\)`)
	reNestMethod     = regexp.MustCompile(`@(?i:(Get|Post|Put|Patch|Delete))\(\s*['"]?([^'"]*)['"]?\s*\)`)

	// FastAPI
	reFastAPI = regexp.MustCompile(`@app\.(get|post|put|patch|delete)\(\s*['"]([^'"]+)['"]`)

	// Django
	reDjangoPath   = regexp.MustCompile(`\bpath\(\s*['"]([^'"]+)['"]`)
	reDjangoRePath = regexp.MustCompile(`\bre_path\(\s*['"]([^'"]+)['"]`)

	// Gin
	reGin = regexp.MustCompile(`\b(?:r|router|\w+)\.(GET|POST|PUT|PATCH|DELETE)\(\s*['"]([^'"]+)['"]`)
	// group := r.Group("/api")
	reGinGroupInit = regexp.MustCompile(`(\w+)\s*:=\s*(?:r|router|\w+)\.Group\(\s*['"]([^'"]+)['"]\s*\)`)
	// sub := group.Group("/v1")
	reGinGroupNest = regexp.MustCompile(`(\w+)\s*:=\s*(\w+)\.Group\(\s*['"]([^'"]+)['"]\s*\)`)
	// group.GET("/x")
	reGinGroupCall = regexp.MustCompile(`(\w+)\.(GET|POST|PUT|PATCH|DELETE)\(\s*['"]([^'"]+)['"]`)

	// Gorilla/mux
	reMuxHandle  = regexp.MustCompile(`\bHandleFunc\(\s*['"]([^'"]+)['"]\s*,`)
	reMuxMethods = regexp.MustCompile(`\.Methods\(\s*['"]?(GET|POST|PUT|PATCH|DELETE)['"]?\s*\)`)

	// Chi
	reChiSimple = regexp.MustCompile(`\b(?:r|router|\w+)\.(Get|Post|Put|Patch|Delete)\(\s*['"]([^'"]+)['"]`)
	// group := chi.NewRouter() / r.Route("/api", func(r chi.Router) { ... })
	reChiRouteBlock = regexp.MustCompile(`\.\s*Route\(\s*['"]([^'"]+)['"]\s*,\s*func\(\s*\w+\s+chi\.Router\)`)

	// Fiber
	reFiber          = regexp.MustCompile(`\b(?:app|router|group|\w+)\.(Get|Post|Put|Patch|Delete)\(\s*['"]([^'"]+)['"]`)
	reFiberGroupInit = regexp.MustCompile(`(\w+)\s*:=\s*(?:app|router|\w+)\.Group\(\s*['"]([^'"]+)['"]\s*\)`)
	reFiberGroupNest = regexp.MustCompile(`(\w+)\s*:=\s*(\w+)\.Group\(\s*['"]([^'"]+)['"]\s*\)`)
	reFiberGroupCall = regexp.MustCompile(`(\w+)\.(Get|Post|Put|Patch|Delete)\(\s*['"]([^'"]+)['"]`)

	// Spring
	reSpringVerb   = regexp.MustCompile(`@(?i:(Get|Post|Put|Patch|Delete))Mapping\(\s*["']([^"']*)["']?`)
	reSpringReqMap = regexp.MustCompile(`@RequestMapping\(\s*["']([^"']*)["']?`)

	// ASP.NET
	reAspVerb  = regexp.MustCompile(`\[(HttpGet|HttpPost|HttpPut|HttpPatch|HttpDelete)(?:\s*\(\s*"([^"]*)"\s*\))?\]`)
	reAspRoute = regexp.MustCompile(`\[Route\(\s*"([^"]+)"\s*\)\]`)

	// Rails
	reRails = regexp.MustCompile(`\b(get|post|put|patch|delete)\s+['"]([^'"]+)['"]`)

	reYamlPathKey = regexp.MustCompile(`(?m)^\s{0,4}(/[^:\s]+):\s*$`)
	reYamlVerbKey = regexp.MustCompile(`(?m)^\s{2,6}(get|post|put|patch|delete):\s*$`)
)

func readRouteFile(full, rel string, routeMaxLines int) RouteFile {
	out := RouteFile{Path: rel, Meta: map[string]string{}}

	b, _ := os.ReadFile(full)
	ext := strings.ToLower(filepath.Ext(rel))
	text := string(b)

	sc := bufio.NewScanner(bytes.NewReader(b))
	seen := 0
	indicators := []string{
		"Route::",
		".get(", ".post(", ".put(", ".patch(", ".delete(",
		".Get(", ".Post(", ".Put(", ".Patch(", ".Delete(",
		"->middleware(", "->name(", "->group(", "prefix(",
		"@Get(", "@Post(", "@Put(", "@Patch(", "@Delete(", "@Controller(",
		"@app.get(", "@app.post(", "@app.put(", "@app.patch(", "@app.delete(",
		"path(", "re_path(", "HandleFunc(", "[HttpGet", "[HttpPost", "[HttpPut", "[HttpPatch", "[HttpDelete", "[Route(",
		"@RequestMapping(", "@GetMapping(", "@PostMapping(", "@PutMapping(", "@PatchMapping(", "@DeleteMapping(",
		"paths:",
	}
	for sc.Scan() {
		line := stripInlineComments(strings.TrimSpace(sc.Text()))
		if line == "" {
			continue
		}
		for _, k := range indicators {
			if strings.Contains(line, k) {
				out.Snips = append(out.Snips, truncate(line, 240))
				seen++
				break
			}
		}
		if seen >= routeMaxLines {
			break
		}
	}

	guessedSet := make(map[string]struct{})

	if ext == ".json" {
		parseOpenAPIJSON(text, &out, guessedSet)
	}
	if ext == ".yaml" || ext == ".yml" {
		parseOpenAPIYAMLHeuristic(text, &out, guessedSet)
	}

	parseLaravel(text, &out, guessedSet)

	parseExpress(text, &out, guessedSet)

	parseNest(text, &out, guessedSet)

	for _, m := range reFastAPI.FindAllStringSubmatch(text, -1) {
		addGuess(guessedSet, &out.Guessed, m[1], m[2])
	}

	for _, m := range reDjangoPath.FindAllStringSubmatch(text, -1) {
		addGuess(guessedSet, &out.Guessed, "ANY", m[1])
	}
	for _, m := range reDjangoRePath.FindAllStringSubmatch(text, -1) {
		addGuess(guessedSet, &out.Guessed, "ANY", m[1])
	}

	parseGin(text, &out, guessedSet)

	parseChi(text, &out, guessedSet)

	parseFiber(text, &out, guessedSet)

	parseMux(text, &out, guessedSet)

	parseSpring(text, &out, guessedSet)

	parseAsp(text, &out, guessedSet)

	for _, m := range reRails.FindAllStringSubmatch(text, -1) {
		addGuess(guessedSet, &out.Guessed, m[1], m[2])
	}

	switch ext {
	case ".ts", ".tsx", ".js", ".jsx":
	default:
	}

	return out
}

func parseLaravel(text string, out *RouteFile, set map[string]struct{}) {
	for _, m := range reLaravelRoute.FindAllStringSubmatch(text, -1) {
		addGuess(set, &out.Guessed, m[1], m[2])
	}
	prefixes := reLaravelGroup1.FindAllStringSubmatch(text, -1)
	prefixes = append(prefixes, reLaravelGroup2.FindAllStringSubmatch(text, -1)...)
	if len(prefixes) > 0 {
		base := prefixes[0][1]
		for _, m := range reLaravelRoute.FindAllStringSubmatch(text, -1) {
			addGuess(set, &out.Guessed, m[1], joinPath(base, m[2]))
		}
	}
}

func parseExpress(text string, out *RouteFile, set map[string]struct{}) {
	routerBase := map[string]string{} // routerVar -> '/api'
	for _, m := range reExpressUse.FindAllStringSubmatch(text, -1) {
		base := m[1]
		rv := m[2]
		routerBase[rv] = base
	}

	for _, m := range reExpressRoute.FindAllStringSubmatch(text, -1) {
		method := m[1]
		path := m[2]
		full := m[0]
		caller := ""
		if idx := strings.Index(full, "."); idx > 0 {
			caller = strings.TrimSpace(full[:idx])
		}
		if base, ok := routerBase[caller]; ok && path != "" && strings.HasPrefix(path, "/") {
			addGuess(set, &out.Guessed, method, joinPath(base, path))
		} else {
			addGuess(set, &out.Guessed, method, path)
		}
	}
}

func parseNest(text string, out *RouteFile, set map[string]struct{}) {
	controllerBases := reNestController.FindAllStringSubmatch(text, -1)
	methods := reNestMethod.FindAllStringSubmatch(text, -1)

	if len(controllerBases) == 0 {
		for _, m := range methods {
			addGuess(set, &out.Guessed, m[1], m[2])
		}
		return
	}

	base := controllerBases[0][1]
	for _, m := range methods {
		addGuess(set, &out.Guessed, m[1], joinPath(base, m[2]))
	}
}

func parseGin(text string, out *RouteFile, set map[string]struct{}) {
	for _, m := range reGin.FindAllStringSubmatch(text, -1) {
		addGuess(set, &out.Guessed, m[1], m[2])
	}

	groupBase := map[string]string{} // var -> base
	for _, m := range reGinGroupInit.FindAllStringSubmatch(text, -1) {
		groupBase[m[1]] = m[2]
	}
	for _, m := range reGinGroupNest.FindAllStringSubmatch(text, -1) {
		parent := groupBase[m[2]]
		groupBase[m[1]] = joinPath(parent, m[3])
	}
	for _, m := range reGinGroupCall.FindAllStringSubmatch(text, -1) {
		g := m[1]
		verb := m[2]
		p := m[3]
		if base, ok := groupBase[g]; ok {
			addGuess(set, &out.Guessed, verb, joinPath(base, p))
		} else {
			addGuess(set, &out.Guessed, verb, p)
		}
	}
}

func parseFiber(text string, out *RouteFile, set map[string]struct{}) {
	for _, m := range reFiber.FindAllStringSubmatch(text, -1) {
		addGuess(set, &out.Guessed, m[1], m[2])
	}
	groupBase := map[string]string{}
	for _, m := range reFiberGroupInit.FindAllStringSubmatch(text, -1) {
		groupBase[m[1]] = m[2]
	}
	for _, m := range reFiberGroupNest.FindAllStringSubmatch(text, -1) {
		parent := groupBase[m[2]]
		groupBase[m[1]] = joinPath(parent, m[3])
	}
	for _, m := range reFiberGroupCall.FindAllStringSubmatch(text, -1) {
		g := m[1]
		verb := m[2]
		p := m[3]
		if base, ok := groupBase[g]; ok {
			addGuess(set, &out.Guessed, verb, joinPath(base, p))
		} else {
			addGuess(set, &out.Guessed, verb, p)
		}
	}
}

func parseChi(text string, out *RouteFile, set map[string]struct{}) {
	for _, m := range reChiSimple.FindAllStringSubmatch(text, -1) {
		addGuess(set, &out.Guessed, m[1], m[2])
	}
	for _, m := range reChiRouteBlock.FindAllStringSubmatch(text, -1) {
		addGuess(set, &out.Guessed, "ANY", m[1])
	}
}

func parseMux(text string, out *RouteFile, set map[string]struct{}) {
	if reMuxHandle.MatchString(text) {
		paths := reMuxHandle.FindAllStringSubmatch(text, -1)
		methods := reMuxMethods.FindAllStringSubmatch(text, -1)
		if len(paths) == len(methods) {
			for i := range paths {
				addGuess(set, &out.Guessed, methods[i][1], paths[i][1])
			}
		} else {
			for _, p := range paths {
				addGuess(set, &out.Guessed, "ANY", p[1])
			}
		}
	}
}

func parseSpring(text string, out *RouteFile, set map[string]struct{}) {
	classReq := reSpringReqMap.FindAllStringSubmatch(text, -1)
	methods := reSpringVerb.FindAllStringSubmatch(text, -1)

	if len(classReq) == 0 {
		for _, m := range methods {
			addGuess(set, &out.Guessed, m[1], m[2])
		}
		return
	}
	base := classReq[0][1]
	for _, m := range methods {
		addGuess(set, &out.Guessed, m[1], joinPath(base, m[2]))
	}
}

func parseAsp(text string, out *RouteFile, set map[string]struct{}) {
	classRoutes := reAspRoute.FindAllStringSubmatch(text, -1)
	methods := reAspVerb.FindAllStringSubmatch(text, -1)

	if len(methods) > 0 {
		var base string
		if len(classRoutes) > 0 {
			base = classRoutes[0][1]
		}
		for _, m := range methods {
			verb := strings.ToUpper(strings.TrimPrefix(m[1], "Http"))
			p := trimQuotes(m[2])
			if p == "" {
				if base != "" {
					addGuess(set, &out.Guessed, verb, joinPath(base, ""))
				} else {
					addGuess(set, &out.Guessed, verb, "/")
				}
			} else if base != "" {
				addGuess(set, &out.Guessed, verb, joinPath(base, p))
			} else {
				addGuess(set, &out.Guessed, verb, p)
			}
		}
	} else {
		for _, m := range classRoutes {
			addGuess(set, &out.Guessed, "ANY", m[1])
		}
	}
}

func parseOpenAPIJSON(text string, out *RouteFile, set map[string]struct{}) {
	var obj map[string]any
	if err := json.Unmarshal([]byte(text), &obj); err != nil {
		return
	}
	paths, _ := obj["paths"].(map[string]any)
	if paths == nil {
		return
	}
	for p, v := range paths {
		ops, _ := v.(map[string]any)
		if ops == nil {
			continue
		}
		for verb := range ops {
			lv := strings.ToLower(verb)
			switch lv {
			case "get", "post", "put", "patch", "delete":
				addGuess(set, &out.Guessed, lv, p)
			}
		}
	}
}

func parseOpenAPIYAMLHeuristic(text string, out *RouteFile, set map[string]struct{}) {
	if !strings.Contains(text, "paths:") {
		return
	}
	lines := strings.Split(text, "\n")
	var curPath string
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if m := reYamlPathKey.FindStringSubmatch(l); len(m) == 2 {
			curPath = m[1]
			continue
		}
		if m := reYamlVerbKey.FindStringSubmatch(l); len(m) == 2 && curPath != "" {
			addGuess(set, &out.Guessed, m[1], curPath)
		}
	}
}
