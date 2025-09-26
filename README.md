ctxgen is a command-line tool that scans any project (multi-language) and generates a context manifest in JSON or NDJSON format.
The manifest provides a structured overview of the codebase: framework, dependencies, environment keys, file inventory, Laravel context (controllers, models, routes, migrations, seeders), Git metadata

The main goal: automatically capture project context so you (or an assistant) can load it and immediately understand the setup, stack, and changes without manually typing everything.


Build from source

git clone https://github.com/your-org/ctxgen.git
cd ctxgen
go build -o ctxgen

go run .

ctxgen -root <project-root> -out .context/manifest.json


-root	Path to project root (default .).
-out	Output file (default: print to stdout).
-project	Project name (optional).
-git	Include Git info (default: true).
-samples	Comma-separated globs of files to embed (e.g. app/Http/Controllers/**.php,routes/api.php).
-max-sample-kb	Max bytes per embedded sample (default 128).
-max-samples	Max number of embedded samples (default 50).
-route-lines	Max lines scanned per route file (default 200).
-include-files	Include files TOC (default true).
-max-files	Limit number of files listed in TOC (default 5000).
-toc-sha1	Include SHA1 checksums (slower).
-ndjson-out	Output fulltext as NDJSON .gz (TOC first record, followed by file contents).

Examples

Generate a manifest for a Laravel project:
ctxgen -root ./laundry-backend -out .context/manifest.json -project "Laundry Backend"

Generate manifest + NDJSON fulltext:
ctxgen -root ./laundry-backend -out .context/manifest.json -ndjson-out .context/fulltext.ndjson.gz

Embed sample controllers and route files:
ctxgen -root ./laundry-backend \
  -out .context/manifest.json \
  -samples "app/Http/Controllers/**.php,routes/api.php"
