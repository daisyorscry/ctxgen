package main

type Manifest struct {
	Project     string `json:"project,omitempty"`
	Root        string `json:"root"`
	GeneratedAt string `json:"generated_at"`
	Framework   string `json:"framework,omitempty"`

	Composer *ComposerInfo `json:"composer,omitempty"` // PHP/Laravel
	Node     *NodeInfo     `json:"node,omitempty"`     // Node/Next.js/React/Expo
	Go       *GoInfo       `json:"go,omitempty"`       // Go
	Python   *PythonInfo   `json:"python,omitempty"`   // Python
	Rust     *RustInfo     `json:"rust,omitempty"`     // Rust
	Java     *JavaInfo     `json:"java,omitempty"`     // Maven/Gradle
	DotNet   *DotNetInfo   `json:"dotnet,omitempty"`   // .NET
	Ruby     *RubyInfo     `json:"ruby,omitempty"`     // Ruby/Bundler
	Dart     *DartInfo     `json:"dart,omitempty"`     // Dart/Flutter
	Swift    *SwiftInfo    `json:"swift,omitempty"`    // Swift/SwiftPM/CocoaPods

	EnvKeys     []string     `json:"env_keys,omitempty"`
	CodeSummary *CodeSummary `json:"code_summary,omitempty"`

	Laravel    *LaravelCtx `json:"laravel,omitempty"`
	Routes     []RouteFile `json:"routes,omitempty"`
	Migrations []string    `json:"migrations,omitempty"`
	Seeders    []string    `json:"seeders,omitempty"`

	Git           *GitInfo        `json:"git,omitempty"`
	CustomSignals map[string]bool `json:"custom_signals,omitempty"`

	FilesTotal int         `json:"files_total,omitempty"`
	Files      []FileEntry `json:"files,omitempty"`

	Samples []SampleFile `json:"samples,omitempty"`
}

type ComposerInfo struct {
	Name         string            `json:"name,omitempty"`
	Type         string            `json:"type,omitempty"`
	Require      map[string]string `json:"require,omitempty"`
	RequireDev   map[string]string `json:"require_dev,omitempty"`
	AutoloadPSR4 map[string]string `json:"autoload_psr4,omitempty"`
}

type NodeInfo struct {
	Name            string            `json:"name,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Next            bool              `json:"nextjs,omitempty"`
	Expo            bool              `json:"expo,omitempty"`
	ReactNative     bool              `json:"react_native,omitempty"`
	Typescript      bool              `json:"typescript,omitempty"`
}

type GoInfo struct {
	Module   string   `json:"module,omitempty"`
	Requires []string `json:"requires,omitempty"`
}

type PythonInfo struct {
	HasPip       bool              `json:"has_pip"`
	HasPoetry    bool              `json:"has_poetry"`
	HasPipenv    bool              `json:"has_pipenv"`
	Requirements map[string]string `json:"requirements,omitempty"`
	PyProject    map[string]any    `json:"pyproject,omitempty"`
}

type RustInfo struct {
	Package   string            `json:"package,omitempty"`
	Edition   string            `json:"edition,omitempty"`
	Deps      map[string]string `json:"deps,omitempty"`
	Workspace bool              `json:"workspace"`
}

type JavaInfo struct {
	BuildTool string            `json:"build_tool,omitempty"`
	GroupID   string            `json:"group_id,omitempty"`
	Artifact  string            `json:"artifact,omitempty"`
	Plugins   []string          `json:"plugins,omitempty"`
	Deps      map[string]string `json:"deps,omitempty"`
}

type DotNetInfo struct {
	Projects []DotNetProject `json:"projects,omitempty"`
}

type DotNetProject struct {
	Path string            `json:"path"`
	SDK  string            `json:"sdk,omitempty"`
	Refs map[string]string `json:"package_refs,omitempty"`
}

type RubyInfo struct {
	Gems map[string]string `json:"gems,omitempty"`
}

type DartInfo struct {
	Name         string            `json:"name,omitempty"`
	Flutter      bool              `json:"flutter,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DevDeps      map[string]string `json:"dev_dependencies,omitempty"`
}

type SwiftInfo struct {
	PackageName   string            `json:"package_name,omitempty"`
	Deps          map[string]string `json:"deps,omitempty"`
	UsesCocoaPods bool              `json:"uses_cocoapods,omitempty"`
}

type CodeSummary struct {
	Langs  map[string]int `json:"langs,omitempty"`
	PHPLOC int            `json:"php_loc"`
	Files  int            `json:"files"`
}

type LaravelCtx struct {
	Controllers []PHPClassFile `json:"controllers,omitempty"`
	Middleware  []PHPClassFile `json:"middleware,omitempty"`
	Models      []PHPClassFile `json:"models,omitempty"`
	Traits      []PHPClassFile `json:"traits,omitempty"`
	Helpers     []PHPClassFile `json:"helpers,omitempty"`
}

type PHPClassFile struct {
	Path      string   `json:"path"`
	Namespace string   `json:"namespace,omitempty"`
	Class     string   `json:"class,omitempty"`
	Methods   []string `json:"methods,omitempty"`
}

type RouteFile struct {
	Path    string            `json:"path"`
	Snips   []string          `json:"snips"`
	Guessed []string          `json:"guessed"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type GitInfo struct {
	Branch  string   `json:"branch,omitempty"`
	Changed []string `json:"changed,omitempty"`
}

type SampleFile struct {
	Path      string `json:"path"`
	Bytes     int    `json:"bytes"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Lang string `json:"lang,omitempty"`
	SHA1 string `json:"sha1,omitempty"`
}

type NDJSONTOC struct {
	Type         string      `json:"type"`
	Project      string      `json:"project,omitempty"`
	Root         string      `json:"root,omitempty"`
	GeneratedAt  string      `json:"generated_at,omitempty"`
	FilesTotal   int         `json:"files_total"`
	Files        []FileEntry `json:"files"`
	ManifestLite any         `json:"manifest_lite,omitempty"`
}

type NDJSONFile struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Lang    string `json:"lang,omitempty"`
	SHA1    string `json:"sha1,omitempty"`
	Content string `json:"content"`
}
