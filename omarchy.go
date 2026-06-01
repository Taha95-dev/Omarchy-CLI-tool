package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"omarchy/pkg/backup"
	"omarchy/pkg/cleanup"
	"omarchy/pkg/config"
	"omarchy/pkg/counter"
	"omarchy/pkg/database"
	"omarchy/pkg/deploy"
	"omarchy/pkg/disk"
	"omarchy/pkg/doctor"
	"omarchy/pkg/find"
	"omarchy/pkg/gitsupport"
	"omarchy/pkg/info"
	"omarchy/pkg/runscripts"
	"omarchy/pkg/support"
	"omarchy/pkg/templates"
	"omarchy/pkg/tree"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

var Version = getVersion()

func main() {
	config := config.LoadConfig()

	// Set defaults: config value, or fallback to hardcoded
	defaultType := config.DefaultType
	if defaultType == "" {
		defaultType = "cli"
	}
	backendType := config.TypeBackend
	if backendType == "" {
		backendType = "node"
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Handle Ctrl+C
	// Simpler signal handling - cancel immediately
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Println("\n⚠️ Interrupt received. Cancelling...")
		cancel()
	}()
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") {
		switch os.Args[1] {
		case "sync":
			handleGitSync(ctx)
			return
		case "doctor":
			doctor.RunDoctor(ctx)
			return
		case "count":
			handleCountCommand(ctx)
			return
		case "tree":
			handleTreeCommand(ctx)
			return

		// Fast Quality-of-Life shortcut: 'omarchy dev' auto-runs current project ecosystem
		case "dev":
			projectType := runscripts.DetectProjectType()
			if projectType != "unknown" {
				if err := runscripts.RunScript(projectType, "dev"); err != nil {
					fmt.Printf("❌ %v\n", err)
				}
			}
			return

		// Upgraded build to fall back to project detection if no custom Go build name is given
		case "build":
			buildName := ""
			if len(os.Args) > 2 {
				buildName = os.Args[2]
			}

			// If they provided a name, assume it's a manual Go target
			if buildName != "" {
				if err := RunGoBuild(buildName); err != nil {
					fmt.Printf("❌ Build failed: %v\n", err)
				}
				return
			}

			// Otherwise, check if it's a C++, Node, or C# project building
			projectType := runscripts.DetectProjectType()
			if projectType != "unknown" && projectType != "go" {
				if err := runscripts.RunScript(projectType, "build"); err != nil {
					fmt.Printf("❌ %v\n", err)
				}
			} else {
				// Default fallback to standard go build logic
				if err := RunGoBuild(""); err != nil {
					fmt.Printf("❌ Build failed: %v\n", err)
				}
			}
			return

		case "save":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy save <template-name>")
				return
			}
			templateName := os.Args[2]
			currentDir, _ := os.Getwd()
			if err := templates.SaveTemplate(templateName, currentDir); err != nil {
				fmt.Printf("❌ Failed to save template: %v\n", err)
			} else {
				fmt.Printf("✅ Saved template: %s\n", templateName)
			}
			return
		case "delete-template", "rm-template":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy delete-template <template-name>")
				return
			}
			templateName := os.Args[2]
			templates.DeleteTemplate(templateName)
			return
		case "list-templates", "ls-templates":
			templates.ListTemplates()
			return
		case "help", "--help", "-h":
			showHelp()
			return
		case "tree-build", "t-build":
			handleTreeBuildCommand()
			return
		case "version", "--version":
			fmt.Printf("Omarchy %s\n", Version) // Fixed formatting string layout
			return
		case "fix-git-home":
			gitsupport.HandleFixGitInHome()
			return
		case "deploy":
			HandleDeployCommand()
			return
		case "update":
			handleUpdateCommand()
			return
		case "db":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy db <init|migrate|seed|reset|status>")
				return
			}
			handleDBCommand()
			return
		case "du":
			handleDiskUsageCommand()
			return
		case "backup":
			handleBackupCommand()
			return
		case "info":
			info.ShowInfo()
			return
		case "cleanup", "c-up":
			handleCleanupCommand()
			return
		case "find", "f":
			handleFindCommand()
			return
		case "config":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy config <--edit|--path|--list>")
				return
			}
			handleConfigCommand()
			return
		case "run":
			if len(os.Args) < 3 {
				projectType := runscripts.DetectProjectType()
				if projectType != "unknown" {
					runscripts.ListScripts(projectType)
				}
				return
			}

			scriptName := os.Args[2]
			projectType := runscripts.DetectProjectType()
			if projectType == "unknown" {
				return
			}

			if err := runscripts.RunScript(projectType, scriptName); err != nil {
				fmt.Printf("❌ %v\n", err)
			}
		default:
			fmt.Printf("❌ Unknown command: %s\n", os.Args[1])
			fmt.Println("Run 'omarchy help' for available commands")
			return
		}
	}
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "--version") {
		fmt.Printf("Omarchy %s\n", Version)
		return
	}
	// Define flags with config defaults
	projectType := flag.String("type", defaultType, "Project type: web, cli, lib, fullstack")
	backendLang := flag.String("backend", backendType, "Backend language: go, node")
	// Define command line flags
	projectName := flag.String("name", "my-project", "Project name")
	withGit := flag.Bool("git", false, "Initialize git repository")
	withReact := flag.Bool("react", false, "Add React frontend")
	withGo := flag.Bool("go", false, "Add Go backend")
	withNode := flag.Bool("node", false, "Add Node.js backend")
	withVue := flag.Bool("vue", false, "Add Vue frontend")
	withSvelte := flag.Bool("svelte", false, "Add Svelte frontend")
	withNext := flag.Bool("next", false, "Add Next.js frontend")
	withPython := flag.Bool("python", false, "Add Python backend")
	withCpp := flag.Bool("cpp", false, "Add C++ backend")

	flag.Parse()

	// Create project root
	rootPath := filepath.Join(".", *projectName)
	os.MkdirAll(rootPath, 0755)

	// Structure based on project type
	switch *projectType {
	case "web":
		createWebStructure(rootPath, *withReact, *withGo)
	case "cli":
		createCLIStructure(rootPath)
	case "fullstack":
		createFullstackStructure(rootPath, *withReact, *withVue, *withSvelte, *withNext, *withGo, *withNode, *withPython, *withCpp)
	case "backend":
		createBackendOnly(rootPath, *backendLang, *withPython, *withCpp)
	default:
		createDefaultStructure(rootPath)
	}
	// Initialize git if requested
	if *withGit {
		initGit(rootPath)
	}
	fmt.Printf("✅ Created %s project: %s\n", *projectType, *projectName)
}
func getVersion() string {
	// Try git first
	if _, err := os.Stat(".git"); err == nil {
		cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	// Fallback to hardcoded (update manually for releases)
	return "v2.4.0"
}
func DockerCleanup(dryRun bool) error {
	if dryRun {
		fmt.Println("🔍 Docker cleanup dry run:")
		cmd := exec.Command("docker", "system", "prune", "--dry-run")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	fmt.Println("🧹 Cleaning up Docker...")
	cmd := exec.Command("docker", "system", "prune", "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func validateConfig(configPath string) {
	fmt.Printf("🔍 Validating config: %s\n", configPath)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ Config file does not exist\n")
		fmt.Printf("   Run 'omarchy config --edit' to create one\n")
		return
	}

	// Read and parse YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ Cannot read config file: %v\n", err)
		return
	}

	// Try to parse as YAML
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("❌ Invalid YAML: %v\n", err)
		fmt.Printf("\n🔧 Common issues:\n")
		fmt.Printf("   - Missing spaces after colons\n")
		fmt.Printf("   - Using tabs instead of spaces\n")
		fmt.Printf("   - Unclosed quotes\n")
		return
	}

	// Validate known fields
	fmt.Printf("✅ YAML syntax is valid\n\n")

	// Check default_type
	if defaultType, ok := cfg["default_type"]; ok {
		validTypes := []string{"web", "cli", "fullstack", "backend"}
		if !contains(validTypes, defaultType.(string)) {
			fmt.Printf("⚠️  Warning: default_type '%s' is not standard\n", defaultType)
			fmt.Printf("   Supported: web, cli, fullstack, backend\n")
		} else {
			fmt.Printf("✅ default_type: %s\n", defaultType)
		}
	} else {
		fmt.Printf("ℹ️  default_type not set (will use 'cli')\n")
	}

	// Check type_backend
	if backendType, ok := cfg["type_backend"]; ok {
		validBackends := []string{"go", "node", "python", "cpp"}
		if !contains(validBackends, backendType.(string)) {
			fmt.Printf("⚠️  Warning: type_backend '%s' is not supported\n", backendType)
			fmt.Printf("   Supported: go, node, python, cpp\n")
		} else {
			fmt.Printf("✅ type_backend: %s\n", backendType)
		}
	} else {
		fmt.Printf("ℹ️  type_backend not set (will use 'node')\n")
	}

	fmt.Printf("\n💡 Config is usable. Run 'omarchy doctor' for full environment check.\n")
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
func handleConfigCommand() {
	if len(os.Args) < 3 {
		fmt.Println("❌ Missing subcommand")
		fmt.Println("Usage: omarchy config --edit")
		fmt.Println("       omarchy config --path")
		fmt.Println("       omarchy config --list")
		return
	}

	configPath := config.GetConfigPath() // You'll need this function

	switch os.Args[2] {
	case "--edit", "-e":
		// Open config in editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Fallback based on OS
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "nano"
			}
		}

		cmd := exec.Command(editor, configPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Failed to open editor: %v\n", err)
			fmt.Printf("   You can manually edit: %s\n", configPath)
			return
		}
		fmt.Printf("✅ Config saved. Run 'omarchy doctor' to validate.\n")

	case "--path", "-p":
		fmt.Println(configPath)

	case "--list", "-l":
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("❌ Failed to read config: %v\n", err)
			return
		}
		fmt.Println(string(data))
	case "--validate", "-v":
		validateConfig(configPath)
	default:
		fmt.Printf("❌ Unknown config subcommand: %s\n", os.Args[2])
		fmt.Println("Available: --edit, --path, --list, --validate")
	}
}
func HandleDeployCommand() {
	deployCmd := flag.NewFlagSet("deploy", flag.ExitOnError)
	platform := deployCmd.String("platform", "render", "Deployment platform")
	auto := deployCmd.Bool("auto", false, "Auto-deploy without prompts")
	deployCmd.Parse(os.Args[2:])

	projectType := deploy.DetectProjectType()
	if projectType == deploy.Unknown {
		fmt.Println("❌ Could not detect project type")
		return
	}
	fmt.Printf("📦 Detected: %s\n", projectType)

	detectedPlatform := deploy.DetectPlatform()
	if *platform != "render" {
		detectedPlatform = deploy.Platform(*platform)
	}

	projectName := filepath.Base(getCurrentDir())

	// Call Deploy with auto flag
	if err := deploy.Deploy(detectedPlatform, projectType, projectName, *auto); err != nil {
		fmt.Printf("❌ Deployment failed: %v\n", err)
	}
}
func handleFindCommand() {
	findCmd := flag.NewFlagSet("find", flag.ExitOnError)
	pattern := findCmd.String("pattern", "", "Search pattern")
	name := findCmd.String("name", "", "Exact filename")
	ext := findCmd.String("ext", "", "File extension")
	ftype := findCmd.String("type", "", "File type (f=file, d=directory)")
	size := findCmd.String("size", "", "Size filter (+10M, -1G, 500K)")
	maxDepth := findCmd.Int("depth", 0, "Max depth")
	caseSensitive := findCmd.Bool("case", false, "Case sensitive")
	useRegex := findCmd.Bool("regex", false, "Use regex pattern")
	verbose := findCmd.Bool("v", false, "Verbose output")
	findCmd.Parse(os.Args[2:])

	path := "."
	if findCmd.NArg() > 0 {
		path = findCmd.Arg(0)
	}

	opts := find.FindOptions{
		Pattern:       *pattern,
		Type:          *ftype,
		Name:          *name,
		Extension:     *ext,
		Size:          *size,
		MaxDepth:      *maxDepth,
		CaseSensitive: *caseSensitive,
		UseRegex:      *useRegex,
	}

	results, err := find.Find(path, opts)
	if err != nil {
		fmt.Printf("❌ Find failed: %v\n", err)
		return
	}

	find.PrintResults(results, *verbose)
}
func handleCleanupCommand() {
	cleanupCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
	dryRun := cleanupCmd.Bool("dry-run", false, "Preview what would be deleted")
	docker := cleanupCmd.Bool("docker", false, "Clean up Docker (dangling images, stopped containers, unused volumes)")
	all := cleanupCmd.Bool("all", false, "Deep clean (node_modules, .cache, etc.)")
	cleanupCmd.Parse(os.Args[2:])

	path := "."
	if cleanupCmd.NArg() > 0 {
		path = cleanupCmd.Arg(0)
	}

	// Expand tilde
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	opts := cleanup.CleanupOptions{
		DryRun: *dryRun,
		All:    *all,
	}

	if *dryRun {
		fmt.Println("🔍 Cleanup dry run - what would be deleted:")
	} else {
		fmt.Println("🧹 Cleaning up...")
	}
	if *docker {
		if err := DockerCleanup(*dryRun); err != nil {
			fmt.Printf("❌ Docker cleanup failed: %v\n", err)
		}
		return
	}
	deleted, totalSize, err := cleanup.RunCleanup(path, opts)
	if err != nil {
		fmt.Printf("❌ Cleanup failed: %v\n", err)
		return
	}

	if len(deleted) == 0 {
		fmt.Println("   Nothing to clean up")
		return
	}

	if *dryRun {
		fmt.Printf("\n📊 Would free: %s\n", cleanup.FormatSize(totalSize))
	} else {
		fmt.Printf("\n✅ Cleanup complete! Freed: %s\n", cleanup.FormatSize(totalSize))
	}
}
func handleBackupCommand() {
	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	dest := backupCmd.String("dest", ".", "Destination directory for backup")
	name := backupCmd.String("name", "", "Custom backup name")
	backupCmd.Parse(os.Args[2:])

	source := "."
	if backupCmd.NArg() > 0 {
		source = backupCmd.Arg(0)
	}

	// Convert to absolute path for better handling
	absSource, err := filepath.Abs(source)
	if err != nil {
		fmt.Printf("❌ Invalid source path: %v\n", err)
		return
	}

	// Check if source exists
	if _, err := os.Stat(absSource); os.IsNotExist(err) {
		fmt.Printf("❌ Source does not exist: %s\n", source)
		fmt.Printf("   Try using absolute path or check the folder name\n")
		return
	}

	opts := backup.BackupOptions{
		Source: absSource,
		Dest:   *dest,
		Name:   *name,
	}

	if err := backup.CreateBackup(opts); err != nil {
		fmt.Printf("❌ Backup failed: %v\n", err)
	}
}
func handleDiskUsageCommand() {
	depth := 2
	path := "."

	// Parse args...

	// First, replace any backslashes with forward slashes for consistent handling
	path = strings.ReplaceAll(path, "\\", "/")

	// Expand tilde
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		rest := strings.TrimPrefix(path, "~")
		path = filepath.Join(home, rest)
	}

	// Convert to system path
	path = filepath.FromSlash(path)

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err == nil {
		path = absPath
	}

	fmt.Printf("DEBUG: Final path = %s\n", path)

	if err := disk.ShowDiskUsage(path, depth, true); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}
}
func handleDBCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: omarchy db <init|migrate|seed|reset|status> [--dry-run] [--force]")
		return
	}

	subCmd := os.Args[2]
	dbType := database.DetectDatabase()

	// Check for flags
	dryRun := false
	force := false
	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--dry-run":
			dryRun = true
		case "--force":
			force = true
		}
	}

	switch subCmd {
	case "init":
		projectName := filepath.Base(getCurrentDir())
		if dryRun {
			fmt.Println("🔍 DRY RUN: Would initialize database")
			fmt.Printf("   Database type: %s\n", dbType)
			fmt.Printf("   Project name: %s\n", projectName)
			return
		}
		if err := database.InitDatabase(dbType, projectName); err != nil {
			fmt.Printf("❌ Failed to init database: %v\n", err)
		} else {
			fmt.Println("✅ Database initialized successfully.")
		}

	case "migrate":
		// Safety: require --force for production
		if !dryRun && !force && isProductionDB(dbType) {
			fmt.Println("⚠⚠⚠ PRODUCTION DATABASE DETECTED ⚠⚠⚠")
			fmt.Printf("You are about to run migrations on: %s\n", getDBURL(dbType))
			fmt.Println("This could change your schema and potentially delete data.")
			fmt.Println("\nFirst, run dry-run to see what will change:")
			fmt.Println("  omarchy db migrate --dry-run")
			fmt.Println("\nIf you're sure, run with --force:")
			fmt.Println("  omarchy db migrate --force")
			return
		}

		if dryRun {
			database.RunMigration(dbType, true)
		} else {
			if err := database.RunMigration(dbType, false); err != nil {
				fmt.Printf("❌ Migration failed: %v\n", err)
			}
		}

	case "reset":
		// AUTO DRY-RUN FIRST (unless --force bypasses everything)
		if !force {
			fmt.Println("⚠ DATABASE RESET DETECTED")
			fmt.Println("Running dry-run preview first...")
			fmt.Println()

			// Auto dry-run (show what will be deleted)
			database.PreviewReset(dbType) // You'll need this function

			fmt.Println()
			fmt.Println("⚠⚠⚠ WARNING ⚠⚠⚠")
			fmt.Printf("You are about to DELETE ALL DATA in: %s\n", getDBURL(dbType))
			fmt.Println("This action CANNOT be undone.")
			fmt.Println()
			fmt.Print("Type 'DELETE' to confirm, 'dry-run' to preview again, or anything else to cancel: ")

			var confirm string
			fmt.Scanln(&confirm)

			if confirm == "dry-run" {
				// Run preview again and re-prompt
				database.PreviewReset(dbType)
				fmt.Print("Type 'DELETE' to confirm: ")
				fmt.Scanln(&confirm)
			}

			if confirm != "DELETE" {
				fmt.Println("Reset cancelled.")
				return
			}

			// Final warning
			fmt.Println()
			fmt.Print("LAST CHANCE: Type 'I UNDERSTAND' to proceed: ")
			var finalConfirm string
			fmt.Scanln(&finalConfirm)
			if finalConfirm != "I UNDERSTAND" {
				fmt.Println("Reset cancelled.")
				return
			}
		}

		// Actually reset
		if err := database.ResetDatabase(dbType); err != nil {
			fmt.Printf("❌ Reset failed: %v\n", err)
		} else {
			fmt.Println("✅ Database reset successfully.")
			if !force {
				fmt.Println("   Tip: Run 'omarchy db seed' to populate with test data.")
			}
		}

	case "seed":
		// Similar safety for seed (if it overwrites data)
		if !dryRun && !force && isProductionDB(dbType) {
			fmt.Println("⚠⚠⚠ PRODUCTION DATABASE DETECTED ⚠⚠⚠")
			fmt.Print("Seeding will add/overwrite data. Type 'SEED' to continue: ")
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "SEED" {
				fmt.Println("Seed cancelled.")
				return
			}
		}

		if dryRun {
			database.SeedDatabase(dbType) // Should show preview
		} else {
			if err := database.SeedDatabase(dbType); err != nil {
				fmt.Printf("❌ Seeding failed: %v\n", err)
			}
		}

	case "status":
		fmt.Printf("📊 Database: %s\n", dbType)
		// Show migration status
	default:
		fmt.Printf("Unknown db command: %s\n", subCmd)
		fmt.Println("Available: init, migrate, seed, reset, status")
	}
}

// Helper functions you'll need to add in pkg/database/

func isProductionDB(dbType database.DatabaseType) bool {
	// Check database URL for production indicators
	_ = dbType
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return false
	}
	// Look for production keywords
	prodIndicators := []string{"prod", "production", "live", "aws", "heroku", "railway"}
	for _, indicator := range prodIndicators {
		if strings.Contains(strings.ToLower(url), indicator) {
			return true
		}
	}
	return false
}

func getDBURL(dbType database.DatabaseType) string {
	_ = dbType
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "localhost (default)"
	}
	// Mask sensitive info (password)
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) > 1 {
			return parts[1] // Show only host part
		}
	}
	return url
}
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "my-project"
	}
	return dir
}
func createBackendOnly(path string, backendType string, withPython, withCpp bool) {
	if withPython {
		createPythonBackend(path)
	} else if withCpp {
		createCppBackend(path)
	} else {
		switch backendType {
		case "go":
			createGoBackend(path)
		case "node":
			createNodeBackend(path)
		default:
			createGoBackend(path)
		}
	}
}
func createPythonBackend(path string) {
	// Create folder structure
	folders := []string{
		"app",
		"app/api",
		"app/models",
		"app/schemas",
		"tests",
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	// requirements.txt
	requirements := `fastapi==0.104.1
uvicorn==0.24.0
sqlalchemy==2.0.23
pydantic==2.5.0
python-dotenv==1.0.0
`
	os.WriteFile(filepath.Join(path, "requirements.txt"), []byte(requirements), 0644)

	// main.py
	mainPy := `from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import os
from dotenv import load_dotenv

load_dotenv()

app = FastAPI(title="Omarchy Python API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "Omarchy Python API"}

@app.get("/health")
def health():
    return {"status": "ok"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
`
	os.WriteFile(filepath.Join(path, "app/main.py"), []byte(mainPy), 0644)

	// .env
	envFile := `DATABASE_URL=sqlite:///./app.db
SECRET_KEY=your-secret-key-here
`
	os.WriteFile(filepath.Join(path, ".env"), []byte(envFile), 0644)

	// Dockerfile (optional)
	dockerfile := `FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
`
	os.WriteFile(filepath.Join(path, "Dockerfile"), []byte(dockerfile), 0644)

	fmt.Println("✅ Created Python FastAPI backend")
}
func createCppBackend(path string) {
	// Create folder structure
	folders := []string{
		"src",
		"include",
		"build",
		"tests",
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	// CMakeLists.txt
	cmake := `cmake_minimum_required(VERSION 3.20)
project(OmarchyCppAPI)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find Crow (web framework)
include(FetchContent)
FetchContent_Declare(
    crow
    GIT_REPOSITORY https://github.com/CrowCpp/Crow.git
    GIT_TAG v1.0.5
)
FetchContent_MakeAvailable(crow)

add_executable(omarchy-api src/main.cpp)
target_link_libraries(omarchy-api Crow::Crow)

# Install
install(TARGETS omarchy-api DESTINATION bin)
`
	os.WriteFile(filepath.Join(path, "CMakeLists.txt"), []byte(cmake), 0644)

	// src/main.cpp
	mainCpp := `#include <crow.h>
#include <iostream>

int main() {
    crow::SimpleApp app;

    CROW_ROUTE(app, "/")([](){
        return crow::response("Omarchy C++ API");
    });

    CROW_ROUTE(app, "/health")([](){
        crow::json::wvalue result;
        result["status"] = "ok";
        return result;
    });

    std::cout << "Server starting on http://localhost:8080" << std::endl;
    app.port(8080).multithreaded().run();

    return 0;
}
`
	os.WriteFile(filepath.Join(path, "src/main.cpp"), []byte(mainCpp), 0644)

	// Dockerfile
	dockerfile := `FROM gcc:latest

WORKDIR /app

RUN apt-get update && apt-get install -y cmake

COPY . .

RUN mkdir build && cd build && cmake .. && make

CMD ["./build/omarchy-api"]
`
	os.WriteFile(filepath.Join(path, "Dockerfile"), []byte(dockerfile), 0644)

	// .gitignore additions for C++
	gitignoreCpp := `build/
*.o
*.exe
`
	f, _ := os.OpenFile(filepath.Join(path, ".gitignore"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString(gitignoreCpp)
	f.Close()

	fmt.Println("✅ Created C++ Crow backend")
}
func createWebStructure(path string, withReact, withGo bool) {
	folders := []string{
		"src/css",
		"src/js",
		"src/assets",
		"public",
	}

	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	if withReact {
		os.MkdirAll(filepath.Join(path, "src/components"), 0755)
		createReactFiles(path)
	}

	if withGo {
		createGoBackend(path)
	}
}

func createCLIStructure(path string) {
	folders := []string{
		"cmd",
		"internal",
		"pkg",
		"scripts",
	}

	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	// Create main.go
	mainContent := `package main

import "fmt"

func main() {
    fmt.Println("Omarchy CLI Tool")
}
`
	os.WriteFile(filepath.Join(path, "main.go"), []byte(mainContent), 0644)
}

func createFullstackStructure(path string, withReact, withVue, withSvelte, withNext, withGo, withNode, withPython, withCpp bool) {
	// Frontend folder (same as before)
	var frontendPath string
	switch {
	case withReact:
		frontendPath = filepath.Join(path, "frontend-react")
		createReactFiles(frontendPath)
	case withVue:
		frontendPath = filepath.Join(path, "frontend-vue")
		createVueFiles(frontendPath)
	case withSvelte:
		frontendPath = filepath.Join(path, "frontend-svelte")
		createSvelteFiles(frontendPath)
	case withNext:
		frontendPath = filepath.Join(path, "frontend-next")
		createNextFiles(frontendPath)
	}

	// Backend folder
	if withGo {
		createGoBackend(filepath.Join(path, "backend-go"))
	}
	if withNode {
		createNodeBackend(filepath.Join(path, "backend-node"))
	}
	if withPython {
		createPythonBackend(filepath.Join(path, "backend-python"))
	}
	if withCpp {
		createCppBackend(filepath.Join(path, "backend-cpp"))
	}
}
func createVueFiles(path string) {
	// Create folders
	folders := []string{
		"src/components",
		"src/views",
		"src/assets",
		"public",
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	// package.json
	packageJSON := `{
  "name": "omarchy-vue-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.0"
  }
}`
	os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJSON), 0644)

	// App.vue
	appVue := `<template>
  <div>
    <h1>Omarchy Vue App</h1>
    <button @click="count++">Count: {{ count }}</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
const count = ref(0)
</script>

<style scoped>
h1 {
  color: #42b883;
}
</style>`
	os.WriteFile(filepath.Join(path, "src/App.vue"), []byte(appVue), 0644)

	// main.js
	mainJS := `import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')`
	os.WriteFile(filepath.Join(path, "src/main.js"), []byte(mainJS), 0644)

	// index.html
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Omarchy Vue App</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>`
	os.WriteFile(filepath.Join(path, "index.html"), []byte(indexHTML), 0644)
}
func createSvelteFiles(path string) {
	folders := []string{
		"src/components",
		"src/lib",
		"public",
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	packageJSON := `{
  "name": "omarchy-svelte-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "svelte": "^4.2.0"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^3.0.0",
    "vite": "^5.0.0"
  }
}`
	os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJSON), 0644)

	appSvelte := `<script>
  let count = 0
</script>

<main>
  <h1>Omarchy Svelte App</h1>
  <button on:click={() => count++}>
    Count: {count}
  </button>
</main>

<style>
  h1 {
    color: #ff3e00;
  }
</style>`
	os.WriteFile(filepath.Join(path, "src/App.svelte"), []byte(appSvelte), 0644)

	mainJS := `import App from './App.svelte'

const app = new App({
  target: document.getElementById('app'),
})

export default app`
	os.WriteFile(filepath.Join(path, "src/main.js"), []byte(mainJS), 0644)

	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Omarchy Svelte App</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>`
	os.WriteFile(filepath.Join(path, "index.html"), []byte(indexHTML), 0644)
}
func createNextFiles(path string) {
	folders := []string{
		"pages",
		"components",
		"styles",
		"public",
	}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	packageJSON := `{
  "name": "omarchy-next-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "next": "^14.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  }
}`
	os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJSON), 0644)

	indexJS := `import Head from 'next/head'
import { useState } from 'react'

export default function Home() {
  const [count, setCount] = useState(0)

  return (
    <>
      <Head>
        <title>Omarchy Next App</title>
      </Head>
      <main>
        <h1>Omarchy Next.js App</h1>
        <button onClick={() => setCount(count + 1)}>
          Count: {count}
        </button>
      </main>
    </>
  )
}`
	os.WriteFile(filepath.Join(path, "pages/index.js"), []byte(indexJS), 0644)
}
func initGit(path string) {
	// Create .gitignore
	gitignore := `# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
*.exe
*.exe~
*.dll
*.so
*.dylib

# IDE
.vscode/
.idea/

# OS
.DS_Store
Thumbs.db
`
	os.WriteFile(filepath.Join(path, ".gitignore"), []byte(gitignore), 0644)

	// Create initial README
	readme := `# Omarchy Project

Created with Omarchy CLI tool.
`
	os.WriteFile(filepath.Join(path, "README.md"), []byte(readme), 0644)
}

func createReactFiles(path string) {
	// package.json
	packageJSON := `{
  "name": "omarchy-frontend",
  "version": "1.0.0",
  "scripts": {
    "start": "vite",
    "build": "vite build"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.0.0",
    "vite": "^4.0.0"
  }
}`
	os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJSON), 0644)

	// App.jsx
	appJSX := `import { useState } from 'react'

function App() {
  const [count, setCount] = useState(0)
  
  return (
    <div>
      <h1>Omarchy React App</h1>
      <button onClick={() => setCount(count + 1)}>
        Count: {count}
      </button>
    </div>
  )
}

export default App
`
	os.WriteFile(filepath.Join(path, "src/App.jsx"), []byte(appJSX), 0644)
}

func createGoBackend(path string) {
	// main.go
	mainGo := `package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, ` + "`" + `{"status": "ok"}` + "`" + `)
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
`
	os.WriteFile(filepath.Join(path, "main.go"), []byte(mainGo), 0644)

	// go.mod
	goMod := `module omarchy-backend

go 1.21
`
	os.WriteFile(filepath.Join(path, "go.mod"), []byte(goMod), 0644)
}

func createDefaultStructure(path string) {
	folders := []string{
		"src",
		"docs",
		"tests",
		"scripts",
	}

	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}
}

// Add to createFullstackStructure or create new function
func createNodeBackend(path string) {
	// Create folder structure
	folders := []string{
		"src/routes",
		"src/controllers",
		"src/models",
		"src/middleware",
		"tests",
	}

	for _, folder := range folders {
		os.MkdirAll(filepath.Join(path, folder), 0755)
	}

	// Create package.json
	packageJSON := `{
  "name": "node-api",
  "version": "1.0.0",
  "scripts": {
    "start": "node src/app.js",
    "dev": "nodemon src/app.js"
  },
  "dependencies": {
    "express": "^4.18.0",
    "cors": "^2.8.5",
    "dotenv": "^16.0.0"
  },
  "devDependencies": {
    "nodemon": "^2.0.0"
  }
}`
	envFile := `PORT=3000
NODE_ENV=development
`
	os.WriteFile(filepath.Join(path, ".env"), []byte(envFile), 0644)
	os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJSON), 0644)

	// Create app.js
	appJS := `const express = require('express')
const cors = require('cors')
require('dotenv').config()

const app = express()
const port = process.env.PORT || 3000

app.use(cors())
app.use(express.json())

app.get('/health', (req, res) => {
    res.json({ status: 'ok' })
})

app.listen(port, () => {
    console.log("Server running on port " + port)
})`
	os.WriteFile(filepath.Join(path, "src/app.js"), []byte(appJS), 0644)
}
func handleGitSync(ctx context.Context) {
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	autoMsg := syncCmd.Bool("a", false, "Auto-generate commit message")
	message := syncCmd.String("m", "", "Commit message")
	dryRun := syncCmd.Bool("dry-run", false, "Show what would happen without doing it")
	tag := syncCmd.String("tag", "", "Create and push tag with this name")

	syncCmd.Parse(os.Args[2:]) // Parse AFTER all flags defined

	if *dryRun {
		gitsupport.DryRunSync(ctx)
		return
	}

	// Run the actual sync — this now completely handles your commit, push, AND tagging safely!
	gitsupport.GitSync(ctx, *autoMsg, *message, *tag)
}
func handleBranchCommand() {
	// Parse flags
	var force bool
	branchName := ""

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-d":
			// Next arg is branch name
			if i+1 < len(os.Args) {
				branchName = os.Args[i+1]
				i++
			} else {
				fmt.Println("❌ Missing branch name")
				return
			}
		case "-D":
			force = true
			if i+1 < len(os.Args) {
				branchName = os.Args[i+1]
				i++
			} else {
				fmt.Println("❌ Missing branch name")
				return
			}
		default:
			// If no flag, assume -d
			if branchName == "" {
				branchName = os.Args[i]
			}
		}
	}

	if branchName == "" {
		fmt.Println("❌ Missing branch name")
		fmt.Println("Usage: omarchy branch -d <name>")
		return
	}

	// Don't delete current branch
	currentBranch := getCurrentBranch()
	if currentBranch == branchName && !force {
		fmt.Printf("❌ Cannot delete branch '%s' while it's checked out.\n", branchName)
		fmt.Println("   Switch to another branch first: git checkout main")
		return
	}

	// If not force, check if branch is merged
	if !force {
		merged, err := isBranchMerged(branchName)
		if err != nil {
			fmt.Printf("❌ Failed to check branch: %v\n", err)
			return
		}
		if !merged {
			fmt.Printf("⚠️ Branch '%s' is not fully merged.\n", branchName)
			fmt.Print("   Use -D to force delete, or merge it first.\n")
			return
		}
	}

	// Confirm deletion
	if !force {
		fmt.Printf("Delete branch '%s'? (y/N): ", branchName)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Delete the branch
	cmd := exec.Command("git", "branch", "-d", branchName)
	if force {
		cmd = exec.Command("git", "branch", "-D", branchName)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to delete branch: %v\n", err)
		fmt.Printf("   %s\n", output)
		return
	}

	fmt.Printf("✅ Deleted branch: %s\n", branchName)
}

// Helper functions
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func isBranchMerged(branchName string) (bool, error) {
	cmd := exec.Command("git", "branch", "--merged", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	// If branch appears in merged list, it's merged
	return strings.Contains(string(output), branchName), nil
}

// When writing files, use consistent line endings
func writeFileContent(path, content string) error {
	// Convert to Unix line endings (LF) for consistency
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return os.WriteFile(path, []byte(content), 0644)
}
func handleCountCommand(ctx context.Context) {
	var recursive bool
	suffix := "go"

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-r", "--recursive":
			recursive = true
		default:
			suffix = os.Args[i]
		}
	}

	var count int
	var err error

	if recursive {
		count, err = counter.CountRecursive(ctx, suffix)
	} else {
		count, err = counter.Count(suffix)
	}

	if err != nil {
		if err == context.Canceled {
			fmt.Println("❌ Operation cancelled by user")
		} else {
			fmt.Printf("❌ Error: %v\n", err)
		}
		return
	}

	if count == 1 {
		fmt.Printf("📁 Found 1 .%s file\n", suffix)
	} else {
		fmt.Printf("📁 Found %d .%s files\n", count, suffix)
	}
}
func handleTreeCommand(ctx context.Context) {
	// Parse depth flag
	depth := 0
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--depth" && i+1 < len(os.Args) {
			fmt.Sscanf(os.Args[i+1], "%d", &depth)
			i++
		}
	}

	root := "."
	if depth > 0 {
		fmt.Printf("📂 Directory tree for: %s (max depth: %d)\n", root, depth)
	} else {
		fmt.Printf("📂 Directory tree for: %s\n", root)
	}
	tree.Print(ctx, root, depth)
}
func RunGoBuild(name string) error {
	outputName := name

	// If no name provided (empty), try to get from go.mod
	if outputName == "" {
		data, err := os.ReadFile("go.mod")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 {
				// Trim spaces AND hidden carriage returns (\r) across all systems
				firstLine := strings.TrimSpace(lines[0])
				if strings.HasPrefix(firstLine, "module ") {
					fullModulePath := strings.TrimSpace(strings.TrimPrefix(firstLine, "module "))

					// Fix: Extract just the last part (e.g., "github.com/user/omarchy" -> "omarchy")
					outputName = filepath.Base(fullModulePath)
				}
			}
		}
	}

	if outputName == "" {
		outputName = "app"
	}

	// Append .exe extension automatically on Windows for clean local running targets
	if runtime.GOOS == "windows" && !strings.HasSuffix(outputName, ".exe") {
		outputName += ".exe"
	}

	// Build the binary locally
	buildCmd := exec.Command("go", "build", "-o", outputName)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		support.PrintError("Build failed. Are you in a Go module directory?")
		support.PrintInfo("Run: go mod init <module-name> first")
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Built binary: %s\n", outputName)

	// Clean Install approach: Tell Go to install the *current folder directory* explicitly
	installCmd := exec.Command("go", "install", ".")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		support.PrintWarning("Install failed, but binary was built locally")
		fmt.Printf("You can still run: ./%s\n", outputName)
		return nil
	}

	fmt.Printf("✅ Installed to GOPATH/bin\n")
	return nil
}
func showHelp() {
	fmt.Printf(`🚀 Omarchy - Project Scaffolding CLI Tool

USAGE:
  omarchy [command] [options]

COMMANDS:
  Project Creation:
    omarchy -name <name> -type <type> [options]   Create new project

  Templates:
    omarchy save <template-name>                  Save current project as template
    omarchy list-templates                        List all saved templates
    omarchy delete-template <name>                Delete a saved template

Database:
   omarchy db init                     Initialize database
   omarchy db migrate                  Run migrations
   omarchy db migrate --dry-run        Preview migrations without running
   omarchy db seed                     Seed database
   omarchy db reset                    Reset database (DESTROYS DATA)
   omarchy db reset --dry-run          Preview reset
   omarchy db status                   Show migration status
   

Git:
  omarchy sync                      Commit with default message
  omarchy sync -a                   Auto-generate commit message
  omarchy sync -m "message"         Commit with custom message
  omarchy sync --tag v2.3.0         Commit + tag + push
  omarchy branch -d <name>          Delete merged branch safely
  omarchy branch -D <name>          Force delete unmerged branch

  Utilities:
    omarchy doctor                               Check development environment
    omarchy count [ext] [-r]                     Count files by extension
    omarchy tree [--depth N]     Show directory tree (limit depth with --depth)
    omarchy version                              Show version
	omarchy tree-build [--preview] [--from file]    Create files/folders from tree structure
	omarchy du [path] [--depth N]    Show disk usage (like du command)
	omarchy backup [path] [--dest DIR] [--name NAME]    Create zip backup of project
	omarchy cleanup [--dry-run] [--all]    Remove temporary files and old backups
	omarchy find [path] --pattern TEXT    Search for files/directories
  --name NAME        Exact filename
  --ext EXT          File extension
  --type f|d         File or directory only
  --size +10M        Size filter (K, M, G)
  --depth N          Max depth
  --regex            Use regex pattern
  --case             Case sensitive
  -v                 Verbose output

  Help:
    omarchy help, omarchy --help                 Show this help message

PROJECT TYPES:
  web          Basic website with HTML/CSS/JS
  cli          Command-line interface tool
  fullstack    Fullstack app with frontend + backend
  backend      Backend only (API server)

FRONTEND OPTIONS (for fullstack/web):
  -react       Add React frontend with Vite
  -vue         Add Vue.js frontend
  -svelte      Add Svelte frontend
  -next        Add Next.js frontend

BACKEND OPTIONS (for fullstack/backend):
  -go          Add Go backend with Gin
  -node        Add Node.js backend with Express
  -python      Add Python backend with FastAPI
  -cpp         Add C++ backend with Crow

OTHER OPTIONS:
  -name <name>          Project name (default: my-project)
  -git                  Initialize git repository
  -type <type>          Project type (default: web)

EXAMPLES:
  # Create a React + Node fullstack app
  omarchy -name my-app -type fullstack -react -node -git

  # Create a Vue + Go fullstack app
  omarchy -name my-app -type fullstack -vue -go

  # Create a backend-only Node API
  omarchy -name my-api -type backend -node

  # Save current project as template
  omarchy save my-starter

  # Auto commit and push all changes
  omarchy sync -a

  # Count all Go files recursively
  omarchy count go -r

  # Auto Update Omarchy CLI tool to latest version
  omarchy update

CONFIGURATION:
  ~/.omarchy.yaml      Default settings (author, license, default_type, etc.)

VERSION:
  Omarchy %s

FEEDBACK:
  Found a bug or have a suggestion? Open an issue on GitHub:
  https://github.com/Taha95-dev/Omarchy-CLI-tool/issues/new

For more information: https://github.com/Taha95-dev/Omarchy-CLI-tool
`, Version)
}
func handleTreeBuildCommand() {
	var fromFile string
	var preview bool

	// Parse flags
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-f", "--from":
			if i+1 < len(os.Args) {
				fromFile = os.Args[i+1]
				i++
			}
		case "--preview":
			preview = true
		}
	}

	var treeText string
	var err error

	if fromFile != "" {
		data, err := os.ReadFile(fromFile)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			return
		}
		treeText = string(data)
	} else {
		fmt.Println("📋 Paste your tree structure (Ctrl+Z then Enter when done):")
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Printf("❌ Failed to read input: %v\n", err)
			return
		}
		treeText = string(data)
	}

	lines := strings.Split(treeText, "\n")
	root, err := tree.ParseTree(lines)
	if err != nil {
		fmt.Printf("❌ Failed to parse tree: %v\n", err)
		return
	}

	if preview {
		fmt.Println("📂 Preview of tree structure:")
		tree.PrintTreePreview(root, "")
		return
	}

	if err := tree.BuildFromTree(root, "."); err != nil {
		fmt.Printf("❌ Failed to build: %v\n", err)
		return
	}

	fmt.Println("\n✅ Tree structure created successfully!")
}
func handleUpdateCommand() {
	fmt.Printf("🔍 Checking for updates...\n")

	// Get latest version from GitHub API
	url := "https://api.github.com/repos/Taha95-dev/Omarchy-CLI-tool/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ Failed to check for updates: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response: %v\n", err)
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	json.Unmarshal(body, &release)

	latestVersion := release.TagName

	if latestVersion == Version {
		fmt.Printf("✅ Already on latest version: %s\n", Version)
		return
	}

	fmt.Printf("📦 New version available: %s (current: %s)\n", latestVersion, Version)

	// Determine binary name based on OS
	var binaryName string
	switch runtime.GOOS {
	case "windows":
		binaryName = "omarchy-windows-amd64.exe"
	case "linux":
		binaryName = "omarchy-linux-amd64"
	case "darwin":
		binaryName = "omarchy-darwin-amd64"
	default:
		fmt.Printf("❌ Unsupported OS: %s\n", runtime.GOOS)
		return
	}

	// Download URL
	downloadURL := fmt.Sprintf("https://github.com/Taha95-dev/Omarchy-CLI-tool/releases/download/%s/%s", latestVersion, binaryName)

	fmt.Printf("⬇️ Downloading update...\n")

	// Download the binary
	resp, err = http.Get(downloadURL)
	if err != nil {
		fmt.Printf("❌ Failed to download: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("❌ Failed to get executable path: %v\n", err)
		return
	}

	// Download to temp file
	tempPath := execPath + ".new"
	out, err := os.Create(tempPath)
	if err != nil {
		fmt.Printf("❌ Failed to create temp file: %v\n", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to save: %v\n", err)
		return
	}
	out.Close()

	// Make executable (Unix)
	if runtime.GOOS != "windows" {
		os.Chmod(tempPath, 0755)
	}

	fmt.Printf("🔄 Installing update...\n")

	// Replace old binary with new one
	err = os.Rename(tempPath, execPath)
	if err != nil {
		// On Windows, you might need to move differently
		fmt.Printf("⚠️ Please run as administrator to complete update\n")
		fmt.Printf("   Or manually replace: %s\n", execPath)
		return
	}

	fmt.Printf("✅ Updated to version %s!\n", latestVersion)
	fmt.Printf("   Run 'omarchy version' to confirm\n")
}
