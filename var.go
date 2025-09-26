package main

import "flag"

var (
	flagRoot = flag.String("root", ".", "project root")
	flagOut  = flag.String("out", "", "output file (default stdout)")
	flagProj = flag.String("project", "", "project name (optional)")
	flagGit  = flag.Bool("git", true, "include git info if available")

	// embed sampel source (optional)
	flagSamplesGlob = flag.String("samples", "", "comma-separated globs to embed (e.g. \"app/Http/Controllers/**.php,routes/api.php\")")
	flagMaxSampleKB = flag.Int("max-sample-kb", 128, "max bytes per embedded sample")
	flagMaxSamples  = flag.Int("max-samples", 50, "hard cap number of embedded samples")

	// route lines/snips
	flagRouteLines = flag.Int("route-lines", 200, "max lines to scan per route file")

	// Files TOC (JSON manifest)
	flagIncludeFiles = flag.Bool("include-files", true, "include files TOC (path/size/lang)")
	flagMaxFiles     = flag.Int("max-files", 5000, "max files listed in TOC")
	flagSHA1         = flag.Bool("toc-sha1", false, "include sha1 for files (slower)")

	// NDJSON GZIP output
	flagNDJSONOut = flag.String("ndjson-out", "", "write fulltext NDJSON GZIP (stream) to this path")
	flagNDJSONExt = flag.String("ndjson-ext",
		"php,js,ts,tsx,jsx,go,py,rb,rs,java,kt,cs,dart,swift,json,yml,yaml,md,sql,xml,gradle,kts,env,sh,txt",
		"comma-separated file extensions to include in NDJSON")
	flagNDJSONSHA1 = flag.Bool("ndjson-sha1", true, "include sha1 in NDJSON records")
)
