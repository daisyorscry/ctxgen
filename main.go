package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	flag.Parse()

	abs, _ := filepath.Abs(*flagRoot)
	now := time.Now().Format(time.RFC3339)

	m := &Manifest{
		Project:       strings.TrimSpace(*flagProj),
		Root:          abs,
		GeneratedAt:   now,
		CustomSignals: map[string]bool{},
	}

	// Composer / Laravel
	m.Composer = readComposer(abs)
	if looksLaravel(m.Composer, abs) {
		m.Framework = "laravel"
	}

	// Node
	m.Node = readPackageJSON(abs)

	// Go
	m.Go = readGoModule(abs)

	// Python / Rust / Java / .NET / Ruby / Dart / Swift
	m.Python = readPython(abs)
	m.Rust = readRust(abs)
	m.Java = readJava(abs)
	m.DotNet = readDotNet(abs)
	m.Ruby = readRuby(abs)
	m.Dart = readDart(abs)
	m.Swift = readSwift(abs)

	// ENV keys
	m.EnvKeys = listEnvKeys(abs)

	// scan project untuk rinkasan & laravel detail & routes/migrations/seeders
	summary, lctx, routes, migrations, seeders := scanProject(abs, *flagRouteLines)
	m.CodeSummary = summary
	m.Laravel = lctx
	m.Routes = routes
	m.Migrations = migrations
	m.Seeders = seeders

	// Files TOC
	if *flagIncludeFiles {
		m.FilesTotal, m.Files = buildFilesTOC(abs, *flagMaxFiles, *flagSHA1)
	}

	// Signals
	m.CustomSignals["jwt"] = hasComposer(m.Composer, "tymon/jwt-auth") || phpHas(lctx, "Jwt") || pkgHas(m.Node, []string{"jsonwebtoken"}) || goHas(m.Go, []string{"github.com/golang-jwt/jwt", "github.com/dgrijalva/jwt-go"})
	m.CustomSignals["otp"] = phpHas(lctx, "Otp") || filesExist(abs, []string{"**/otp/**", "**/*Otp*.php"})
	m.CustomSignals["queue"] = hasComposer(m.Composer, "laravel/horizon") || nodeHasScript(m.Node, "worker") || goHas(m.Go, []string{"github.com/rabbitmq/amqp091-go"}) || filesExist(abs, []string{"**/queue/**"})
	m.CustomSignals["redis"] = hasComposer(m.Composer, "predis/predis") || nodeDepsHas(m.Node, "ioredis") || goHas(m.Go, []string{"github.com/redis/go-redis"})

	// Git
	if *flagGit {
		m.Git = readGit(abs)
	}

	if strings.TrimSpace(*flagSamplesGlob) != "" {
		m.Samples = embedSamples(abs, *flagSamplesGlob, *flagMaxSampleKB*1024, *flagMaxSamples)
	}

	if strings.TrimSpace(*flagNDJSONOut) != "" {
		if err := writeNDJSON(abs, m, *flagNDJSONOut, *flagNDJSONExt, *flagNDJSONSHA1); err != nil {
			fmt.Fprintf(os.Stderr, "ndjson error: %v\n", err)
			os.Exit(1)
		}
	}

	if strings.TrimSpace(*flagOut) != "" {
		outJSON(*flagOut, m)
		return
	}

	// default: print manifest JSON ke stdout
	outJSON("", m)
}
