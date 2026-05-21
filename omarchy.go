package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	config := loadConfig()

	// Set defaults: config value, or fallback to hardcoded
	defaultType := config.DefaultType
	if defaultType == "" {
		defaultType = "cli"
	}
	backendType := config.TypeBackend
	if backendType == "" {
		backendType = "node"
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "sync":
			handleGitSync()
			return
		case "doctor":
			runDoctor()
			return
		case "count":
			handleCountCommand()
			return
		case "version", "--version":
			fmt.Println("Omarchy v0.2.0")
			return
		}
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
		createFullstackStructure(rootPath, *withReact, *withGo, *withNode)
	case "backend":
		createBackendOnly(rootPath, *backendLang)
	default:
		createDefaultStructure(rootPath)
	}
	// Initialize git if requested
	if *withGit {
		initGit(rootPath)
	}
	fmt.Printf("✅ Created %s project: %s\n", *projectType, *projectName)
}
func createBackendOnly(path string, backendType string) {
	switch backendType {
	case "go":
		createGoBackend(path)
	case "node":
		createNodeBackend(path)
	default:
		createGoBackend(path)
	}
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

func createFullstackStructure(path string, withReact, withGo, withNode bool) {
	// Frontend folder
	os.MkdirAll(filepath.Join(path, "frontend/src/components"), 0755)
	os.MkdirAll(filepath.Join(path, "frontend/public"), 0755)

	// Backend folder
	os.MkdirAll(filepath.Join(path, "backend/cmd"), 0755)
	os.MkdirAll(filepath.Join(path, "backend/internal"), 0755)
	os.MkdirAll(filepath.Join(path, "backend/pkg"), 0755)

	// Shared
	os.MkdirAll(filepath.Join(path, "shared"), 0755)
	os.MkdirAll(filepath.Join(path, "scripts"), 0755)

	if withReact {
		createReactFiles(filepath.Join(path, "frontend"))
	}

	if withGo {
		createGoBackend(filepath.Join(path, "backend"))
	}
	if withNode {
		createNodeBackend(filepath.Join(path, "backend"))
	}
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

type Config struct {
	Author       string `yaml:"author"`
	License      string `yaml:"license"`
	DefaultType  string `yaml:"default_type"`
	DefaultReact bool   `yaml:"default_react"`
	DefaultGit   bool   `yaml:"default_git"`
	TypeBackend  string `yaml:"default_backend"`
	Description  string `yaml:"default_description"`
}

func loadConfig() Config {
	var config Config

	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".omarchy.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config // Return empty config (zero values)
	}

	yaml.Unmarshal(data, &config)
	fmt.Println("✅ Config loaded from:", configPath)
	return config
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
func gitSync(autoMsg bool, customMsg string) {
	if !isGitRepo() {
		fmt.Println("❌ Not a git repository")
		return
	}

	if !hasGitChanges(".") {
		fmt.Println("📝 No changes to commit")
		return
	}

	// Add changes
	runCmd("git", "add", ".")

	// Create commit message
	message := customMsg
	if message == "" && autoMsg {
		message = fmt.Sprintf("Auto-sync: %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	if message == "" && !autoMsg {
		message = "Sync via Omarchy"
	}

	// Commit
	runCmd("git", "commit", "-m", message)
	fmt.Println("✅ Committed:", message)

	// Check if remote exists before pushing
	if hasRemote() {
		runCmd("git", "push")
		fmt.Println("✅ Pushed to remote")
	} else {
		fmt.Println("⚠️ No remote configured. Commit saved locally only.")
		fmt.Println("   Run: git remote add origin <url>")
	}
}

func hasRemote() bool {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	err := cmd.Run()
	return err == nil
}
func isGitRepo() bool {
	return exec.Command("git", "-C", ".", "rev-parse", "--is-inside-work-tree").Run() == nil
}
func hasGitChanges(path string) bool {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(output) > 0
}

func handleGitSync() {
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	autoMsg := syncCmd.Bool("a", false, "Auto-generate commit message")
	message := syncCmd.String("m", "", "Commit message")
	syncCmd.Parse(os.Args[2:])

	gitSync(*autoMsg, *message)
}
func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("⚠️ Warning:", err)
	}
}
func runDoctor() {
	fmt.Println("🔍 Omarchy Environment Check")

	// Check tools
	CheckTool("node", "--version")
	CheckTool("go", "version")
	CheckTool("git", "--version")
	CheckTool("npm", "--version")

	// Check config
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fmt.Println("✅ Omarchy config found at:", configPath)
	} else {
		fmt.Println("⚠️ Omarchy config not found (run: omarchy config init)")
	}

	// Check Git config
	fmt.Println("\n📋 Git Configuration:")
	checkGitConfig()
}
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".omarchy.yaml")
}
func checkGitConfig() {
	// Check if git user.name is set
	cmd := exec.Command("git", "config", "--global", "user.name")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("⚠️ Git user.name not set (run: git config --global user.name \"Your Name\")")
	} else {
		fmt.Println("✅ Git user.name:", string(output))
	}

	// Check if git user.email is set
	cmd = exec.Command("git", "config", "--global", "user.email")
	output, err = cmd.Output()
	if err != nil {
		fmt.Println("⚠️ Git user.email not set (run: git config --global user.email \"you@example.com\")")
	} else {
		fmt.Println("✅ Git user.email:", string(output))
	}
}
func checkGitRemote() {
	// Check if remote origin exists
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("⚠️ No git remote configured (run: git remote add origin <url>)")
	} else {
		fmt.Println("✅ Remote origin:", string(output))
	}
}
func CheckTool(tool string, versionFlag string) {
	cmd := exec.Command(tool, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ %s is not installed or not in PATH\n", tool)
	} else {
		// Extract version number
		version := strings.TrimSpace(string(output))
		fmt.Printf("✅ %s %s\n", tool, version)
	}
}
func countAllFiles(suffix string) (int, error) {
	// Remove leading dot if user added it
	suffix = strings.TrimPrefix(suffix, ".")

	entries, err := os.ReadDir(".")
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "."+suffix) {
			count++
		}
	}

	return count, nil
}
func handleCountCommand() {
	// Default to "go" if no suffix provided
	suffix := "go"
	if len(os.Args) > 2 {
		suffix = os.Args[2]
	}

	count, err := countAllFiles(suffix)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	// Pluralization
	if count == 1 {
		fmt.Printf("📁 Found 1 .%s file\n", suffix)
	} else {
		fmt.Printf("📁 Found %d .%s files\n", count, suffix)
	}
}
func saveTemplates(name string, content string) {
	//future implementation (e.g. save to ~/.omarchy/templates/)
	//someday
}
