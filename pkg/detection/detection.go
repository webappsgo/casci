package detection

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/casapps/casci/pkg/pipeline"
)

// ProjectType represents the detected project type
type ProjectType struct {
	Language        string
	LanguageVersion string
	Framework       string
	BuildTool       string
	PackageManager  string
	TestFramework   string
	HasDockerfile   bool
	HasDockerCompose bool
}

// ProjectConfig represents the complete project configuration
type ProjectConfig struct {
	Type           *ProjectType
	BuildCommands  []string
	TestCommands   []string
	ContainerImage string
	Environment    map[string]string
	Pipeline       *pipeline.Pipeline
	PipelineFormat pipeline.PipelineFormat
}

// Detector handles project type detection
type Detector struct {
	parser *pipeline.Parser
}

// NewDetector creates a new project detector
func NewDetector() *Detector {
	return &Detector{
		parser: pipeline.NewParser(),
	}
}

// Detect analyzes a repository and detects project type
func (d *Detector) Detect(repoPath string) (*ProjectConfig, error) {
	projectType := &ProjectType{}

	// Detect language
	lang, version := d.detectLanguage(repoPath)
	projectType.Language = lang
	projectType.LanguageVersion = version

	// Detect framework
	projectType.Framework = d.detectFramework(repoPath, lang)

	// Detect build tool
	projectType.BuildTool = d.detectBuildTool(repoPath, lang)

	// Detect package manager
	projectType.PackageManager = d.detectPackageManager(repoPath, lang)

	// Detect test framework
	projectType.TestFramework = d.detectTestFramework(repoPath, lang)

	// Check for Docker
	projectType.HasDockerfile = d.fileExists(repoPath, "Dockerfile")
	projectType.HasDockerCompose = d.fileExists(repoPath, "docker-compose.yml") ||
		d.fileExists(repoPath, "docker-compose.yaml")

	// Detect and parse existing pipeline
	pipelineFormat := d.parser.DetectFormat(repoPath)
	var parsedPipeline *pipeline.Pipeline
	if pipelineFormat != pipeline.FormatUnknown {
		log.Printf("Detected existing pipeline format: %s", pipelineFormat)
		var err error
		parsedPipeline, err = d.parser.Parse(repoPath)
		if err != nil {
			log.Printf("Warning: failed to parse pipeline: %v", err)
		}
	}

	// Generate build configuration
	config := &ProjectConfig{
		Type:           projectType,
		BuildCommands:  d.generateBuildCommands(projectType),
		TestCommands:   d.generateTestCommands(projectType),
		ContainerImage: d.selectContainerImage(projectType),
		Environment:    d.generateEnvironment(projectType),
		Pipeline:       parsedPipeline,
		PipelineFormat: pipelineFormat,
	}

	return config, nil
}

// detectLanguage detects the primary programming language
func (d *Detector) detectLanguage(repoPath string) (string, string) {
	// Go
	if d.fileExists(repoPath, "go.mod") {
		version := d.extractGoVersion(repoPath)
		return "go", version
	}

	// Node.js / JavaScript
	if d.fileExists(repoPath, "package.json") {
		version := d.extractNodeVersion(repoPath)
		return "javascript", version
	}

	// Python
	if d.fileExists(repoPath, "requirements.txt") ||
		d.fileExists(repoPath, "Pipfile") ||
		d.fileExists(repoPath, "pyproject.toml") ||
		d.fileExists(repoPath, "setup.py") {
		version := d.extractPythonVersion(repoPath)
		return "python", version
	}

	// Java
	if d.fileExists(repoPath, "pom.xml") {
		version := d.extractJavaVersionFromPom(repoPath)
		return "java", version
	}

	if d.fileExists(repoPath, "build.gradle") ||
		d.fileExists(repoPath, "build.gradle.kts") {
		version := d.extractJavaVersionFromGradle(repoPath)
		return "java", version
	}

	// Ruby
	if d.fileExists(repoPath, "Gemfile") {
		return "ruby", ""
	}

	// Rust
	if d.fileExists(repoPath, "Cargo.toml") {
		return "rust", ""
	}

	// C/C++
	if d.fileExists(repoPath, "CMakeLists.txt") ||
		d.fileExists(repoPath, "Makefile") {
		return "c/c++", ""
	}

	// C#
	if d.hasFilesWithExtension(repoPath, ".csproj") ||
		d.hasFilesWithExtension(repoPath, ".sln") {
		return "csharp", ""
	}

	// PHP
	if d.fileExists(repoPath, "composer.json") {
		return "php", ""
	}

	return "unknown", ""
}

// detectFramework detects the framework being used
func (d *Detector) detectFramework(repoPath, language string) string {
	switch language {
	case "javascript":
		// React
		if d.packageJSONContains(repoPath, "react") {
			return "react"
		}
		// Vue
		if d.packageJSONContains(repoPath, "vue") {
			return "vue"
		}
		// Angular
		if d.fileExists(repoPath, "angular.json") {
			return "angular"
		}
		// Next.js
		if d.packageJSONContains(repoPath, "next") {
			return "nextjs"
		}
		// Express
		if d.packageJSONContains(repoPath, "express") {
			return "express"
		}

	case "python":
		// Django
		if d.fileExists(repoPath, "manage.py") {
			return "django"
		}
		// Flask
		if d.hasImport(repoPath, "flask", ".py") {
			return "flask"
		}
		// FastAPI
		if d.hasImport(repoPath, "fastapi", ".py") {
			return "fastapi"
		}

	case "java":
		// Spring Boot
		if d.pomXMLContains(repoPath, "spring-boot") {
			return "spring-boot"
		}

	case "go":
		// Gin
		if d.goModContains(repoPath, "gin-gonic/gin") {
			return "gin"
		}
		// Echo
		if d.goModContains(repoPath, "labstack/echo") {
			return "echo"
		}
	}

	return ""
}

// detectBuildTool detects the build tool
func (d *Detector) detectBuildTool(repoPath, language string) string {
	switch language {
	case "go":
		return "go"
	case "javascript":
		if d.fileExists(repoPath, "package-lock.json") {
			return "npm"
		}
		if d.fileExists(repoPath, "yarn.lock") {
			return "yarn"
		}
		if d.fileExists(repoPath, "pnpm-lock.yaml") {
			return "pnpm"
		}
		return "npm"
	case "java":
		if d.fileExists(repoPath, "pom.xml") {
			return "maven"
		}
		if d.fileExists(repoPath, "build.gradle") || d.fileExists(repoPath, "build.gradle.kts") {
			return "gradle"
		}
	case "python":
		if d.fileExists(repoPath, "Pipfile") {
			return "pipenv"
		}
		if d.fileExists(repoPath, "poetry.lock") {
			return "poetry"
		}
		return "pip"
	case "rust":
		return "cargo"
	case "ruby":
		return "bundler"
	case "php":
		return "composer"
	}
	return ""
}

// detectPackageManager detects the package manager
func (d *Detector) detectPackageManager(repoPath, language string) string {
	return d.detectBuildTool(repoPath, language) // Often the same
}

// detectTestFramework detects the testing framework
func (d *Detector) detectTestFramework(repoPath, language string) string {
	switch language {
	case "go":
		return "go test"
	case "javascript":
		if d.packageJSONContains(repoPath, "jest") {
			return "jest"
		}
		if d.packageJSONContains(repoPath, "mocha") {
			return "mocha"
		}
		if d.packageJSONContains(repoPath, "vitest") {
			return "vitest"
		}
	case "python":
		if d.fileContains(repoPath, "requirements.txt", "pytest") {
			return "pytest"
		}
		return "unittest"
	case "java":
		if d.pomXMLContains(repoPath, "junit") {
			return "junit"
		}
	}
	return ""
}

// generateBuildCommands generates build commands based on project type
func (d *Detector) generateBuildCommands(pt *ProjectType) []string {
	var commands []string

	switch pt.Language {
	case "go":
		commands = append(commands, "go build -v ./...")
	case "javascript":
		if pt.PackageManager == "npm" {
			commands = append(commands, "npm ci", "npm run build")
		} else if pt.PackageManager == "yarn" {
			commands = append(commands, "yarn install --frozen-lockfile", "yarn build")
		} else if pt.PackageManager == "pnpm" {
			commands = append(commands, "pnpm install --frozen-lockfile", "pnpm build")
		}
	case "python":
		if pt.PackageManager == "pip" {
			commands = append(commands, "pip install -r requirements.txt")
		} else if pt.PackageManager == "pipenv" {
			commands = append(commands, "pipenv install")
		} else if pt.PackageManager == "poetry" {
			commands = append(commands, "poetry install")
		}
	case "java":
		if pt.BuildTool == "maven" {
			commands = append(commands, "mvn clean package -DskipTests")
		} else if pt.BuildTool == "gradle" {
			commands = append(commands, "./gradlew build -x test")
		}
	case "rust":
		commands = append(commands, "cargo build --release")
	case "ruby":
		commands = append(commands, "bundle install")
	}

	return commands
}

// generateTestCommands generates test commands
func (d *Detector) generateTestCommands(pt *ProjectType) []string {
	var commands []string

	switch pt.Language {
	case "go":
		commands = append(commands, "go test -v ./...")
	case "javascript":
		if pt.TestFramework == "jest" {
			commands = append(commands, "npm test")
		} else if pt.TestFramework == "mocha" {
			commands = append(commands, "npm test")
		}
	case "python":
		if pt.TestFramework == "pytest" {
			commands = append(commands, "pytest")
		} else {
			commands = append(commands, "python -m unittest discover")
		}
	case "java":
		if pt.BuildTool == "maven" {
			commands = append(commands, "mvn test")
		} else if pt.BuildTool == "gradle" {
			commands = append(commands, "./gradlew test")
		}
	case "rust":
		commands = append(commands, "cargo test")
	case "ruby":
		commands = append(commands, "bundle exec rspec")
	}

	return commands
}

// selectContainerImage selects the appropriate container image
func (d *Detector) selectContainerImage(pt *ProjectType) string {
	if pt.HasDockerfile {
		return "" // Use Dockerfile
	}

	switch pt.Language {
	case "go":
		if pt.LanguageVersion != "" {
			return "golang:" + pt.LanguageVersion + "-alpine"
		}
		return "golang:1.21-alpine"
	case "javascript":
		if pt.LanguageVersion != "" {
			return "node:" + pt.LanguageVersion + "-alpine"
		}
		return "node:18-alpine"
	case "python":
		if pt.LanguageVersion != "" {
			return "python:" + pt.LanguageVersion + "-slim"
		}
		return "python:3.11-slim"
	case "java":
		return "maven:3.9-eclipse-temurin-17"
	case "rust":
		return "rust:1.75-slim"
	case "ruby":
		return "ruby:3.2-alpine"
	}

	return "ubuntu:22.04" // Default fallback
}

// generateEnvironment generates environment variables
func (d *Detector) generateEnvironment(pt *ProjectType) map[string]string {
	env := make(map[string]string)

	env["CI"] = "true"
	env["CASCI"] = "true"

	switch pt.Language {
	case "go":
		env["CGO_ENABLED"] = "0"
		env["GOOS"] = "linux"
	case "javascript":
		env["NODE_ENV"] = "production"
		env["CI"] = "true"
	case "python":
		env["PYTHONUNBUFFERED"] = "1"
	}

	return env
}

// Helper functions

func (d *Detector) fileExists(repoPath, filename string) bool {
	_, err := os.Stat(filepath.Join(repoPath, filename))
	return err == nil
}

func (d *Detector) hasFilesWithExtension(repoPath, ext string) bool {
	matches, _ := filepath.Glob(filepath.Join(repoPath, "*"+ext))
	return len(matches) > 0
}

func (d *Detector) packageJSONContains(repoPath, pkg string) bool {
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return false
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return false
	}

	if deps, ok := packageJSON["dependencies"].(map[string]interface{}); ok {
		if _, exists := deps[pkg]; exists {
			return true
		}
	}

	if devDeps, ok := packageJSON["devDependencies"].(map[string]interface{}); ok {
		if _, exists := devDeps[pkg]; exists {
			return true
		}
	}

	return false
}

func (d *Detector) fileContains(repoPath, filename, search string) bool {
	data, err := os.ReadFile(filepath.Join(repoPath, filename))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), search)
}

func (d *Detector) pomXMLContains(repoPath, search string) bool {
	return d.fileContains(repoPath, "pom.xml", search)
}

func (d *Detector) goModContains(repoPath, search string) bool {
	return d.fileContains(repoPath, "go.mod", search)
}

func (d *Detector) hasImport(repoPath, importName, ext string) bool {
	// Simple check - could be more sophisticated
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ext) {
			data, _ := os.ReadFile(path)
			if strings.Contains(string(data), "import "+importName) ||
				strings.Contains(string(data), "from "+importName) {
				return filepath.SkipAll
			}
		}
		return nil
	})
	return err == filepath.SkipAll
}

// Version extraction functions

func (d *Detector) extractGoVersion(repoPath string) string {
	data, err := os.ReadFile(filepath.Join(repoPath, "go.mod"))
	if err != nil {
		return "1.21"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "go ") {
			version := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "go "))
			return version
		}
	}
	return "1.21"
}

func (d *Detector) extractNodeVersion(repoPath string) string {
	// Check .nvmrc
	if data, err := os.ReadFile(filepath.Join(repoPath, ".nvmrc")); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Check package.json engines
	data, err := os.ReadFile(filepath.Join(repoPath, "package.json"))
	if err != nil {
		return "18"
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(data, &packageJSON); err != nil {
		return "18"
	}

	if engines, ok := packageJSON["engines"].(map[string]interface{}); ok {
		if node, ok := engines["node"].(string); ok {
			return strings.TrimPrefix(node, "^")
		}
	}

	return "18"
}

func (d *Detector) extractPythonVersion(repoPath string) string {
	// Check .python-version
	if data, err := os.ReadFile(filepath.Join(repoPath, ".python-version")); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Check runtime.txt (Heroku style)
	if data, err := os.ReadFile(filepath.Join(repoPath, "runtime.txt")); err == nil {
		version := strings.TrimPrefix(strings.TrimSpace(string(data)), "python-")
		return version
	}

	return "3.11"
}

func (d *Detector) extractJavaVersionFromPom(repoPath string) string {
	data, err := os.ReadFile(filepath.Join(repoPath, "pom.xml"))
	if err != nil {
		return "17"
	}

	content := string(data)
	if strings.Contains(content, "<java.version>17</java.version>") {
		return "17"
	}
	if strings.Contains(content, "<maven.compiler.target>17</maven.compiler.target>") {
		return "17"
	}

	return "17"
}

func (d *Detector) extractJavaVersionFromGradle(repoPath string) string {
	return "17" // Default for now
}
