package main

import (
	"context"
	"flag"
	"fmt"
	"omarchy/pkg/counter"
	"omarchy/pkg/gitsupport"
	"omarchy/pkg/tree"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var Version = "v1.10.0"

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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "sync":
			handleGitSync(ctx)
			return
		case "doctor":
			runDoctor(ctx)
			return
		case "count":
			handleCountCommand(ctx)
			return
		case "tree":
			handleTreeCommand(ctx)
			return
		case "save":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy save <template-name>")
				return
			}
			templateName := os.Args[2]
			currentDir, _ := os.Getwd()
			if err := saveTemplate(templateName, currentDir); err != nil {
				fmt.Printf("❌ Failed to save template: %v\n", err)
			} else {
				fmt.Printf("✅ Saved template: %s\n", templateName)
			}
			return
		case "delete-template", "rm-template":
			if len(os.Args) < 3 {
				fmt.Println("Usage: omarchy delete-template <template-name>")
				fmt.Println("   or: omarchy rm-template <template-name>")
				return
			}
			templateName := os.Args[2]
			deleteTemplate(templateName)
			return
		case "list-templates", "ls-templates":
			listTemplates()
			return
		case "build":
			buildName := ""
			if len(os.Args) > 2 {
				buildName = os.Args[2]
			}
			if err := RunGoBuild(buildName); err != nil {
				fmt.Printf("❌ Build failed: %v\n", err)
			}
			return
		case "help", "--help", "-h":
			showHelp()
			return
		default:
			fmt.Printf("❌ Unknown command: %s\n", os.Args[1])
			fmt.Println("Run 'omarchy help' for available commands")
			return
		case "version", "--version":
			fmt.Println("Omarchy" + Version)
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
		createFullstackStructure(rootPath, *withReact, *withVue, *withSvelte, *withNext, *withGo, *withNode)
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

func createFullstackStructure(path string, withReact, withVue, withSvelte, withNext, withGo, withNode bool) {
	// Frontend folder
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
func handleGitSync(ctx context.Context) {
	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	autoMsg := syncCmd.Bool("a", false, "Auto-generate commit message")
	message := syncCmd.String("m", "", "Commit message")
	dryRun := syncCmd.Bool("dry-run", false, "Show what would happen without doing it")
	syncCmd.Parse(os.Args[2:])

	if *dryRun {
		gitsupport.DryRunSync(ctx)
		return
	}

	gitsupport.GitSync(ctx, *autoMsg, *message)
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
func runDoctor(ctx context.Context) {
	select {
	case <-ctx.Done():
		fmt.Println("❌ Doctor check cancelled")
		return
	default:
	}

	fmt.Printf("🖥️ Operating System: %s\n", getOS())
	fmt.Printf("🏠 Home Directory: %s\n", getHomeDir())
	fmt.Println("🔍 Omarchy Environment Check")

	// Check tools (these could also be made cancellable)
	CheckTool("node", "--version")
	CheckTool("go", "version")
	CheckTool("git", "--version")
	CheckTool("npm", "--version")
	CheckTool("docker", "--version")

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
func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".omarchy.yaml")
}
func checkGitConfig() (string, string, error) {
	nameCmd := exec.Command("git", "config", "--global", "user.name")
	nameOutput, nameErr := nameCmd.Output()

	emailCmd := exec.Command("git", "config", "--global", "user.email")
	emailOutput, emailErr := emailCmd.Output()

	if nameErr != nil || emailErr != nil {
		fmt.Println("⚠️ Git user.name or user.email not set")
		fmt.Println("   Run: git config --global user.name \"Your Name\"")
		fmt.Println("   Run: git config --global user.email \"you@example.com\"")
		return "", "", fmt.Errorf("git config not set")
	}

	userName := strings.TrimSpace(string(nameOutput))
	userEmail := strings.TrimSpace(string(emailOutput))
	fmt.Printf("✅ Git user: %s <%s>\n", userName, userEmail)
	return userName, userEmail, nil
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
	cmdName := tool
	displayName := tool

	// OS-specific adjustments
	switch runtime.GOOS {
	case "windows":
		switch tool {
		case "node":
			cmdName = "node.exe"
		case "python":
			cmdName = "python.exe"
		case "go":
			cmdName = "go.exe"
		}
	case "darwin": // macOS
		// Usually fine
	case "linux":
		// Usually fine
	}

	cmd := exec.Command(cmdName, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if tool == "docker" {
			fmt.Printf("⚠️ %s not installed (optional)\n", displayName)
		} else {
			fmt.Printf("❌ %s not installed or not in PATH\n", displayName)
		}
	} else {
		version := strings.TrimSpace(string(output))
		// Clean up version string (remove extra spaces, newlines)
		version = strings.Split(version, "\n")[0]
		fmt.Printf("✅ %s %s\n", displayName, version)
	}
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
	root := "."
	fmt.Printf("📂 Directory tree for: %s\n", root)
	tree.Print(ctx, root)
}
func getOS() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		return runtime.GOOS
	}
}
func ConfigPath() string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", ".omarchy.yaml")
	case "darwin": // macOS
		return filepath.Join(home, "Library", "Application Support", "omarchy.yaml")
	default: // Linux and others
		return filepath.Join(home, ".config", "omarchy.yaml")
	}
}
func printError(msg string) {
	fmt.Println("❌", msg)
}

func printWarning(msg string) {
	fmt.Println("⚠️", msg)
}

func printSuccess(msg string) {
	fmt.Println("✅", msg)
}

func printInfo(msg string) {
	fmt.Println("✅", msg)
}
func saveTemplate(name string, sourcePath string) error {
	// Create templates directory
	homeDir, _ := os.UserHomeDir()
	templatesDir := filepath.Join(homeDir, ".omarchy", "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create template folder
	templatePath := filepath.Join(templatesDir, name)
	if _, err := os.Stat(templatePath); err == nil {
		fmt.Printf("⚠️ Template '%s' already exists. Overwrite? (y/N): ", name)
		var resp string
		fmt.Scanln(&resp)
		if resp != "y" && resp != "Y" {
			fmt.Println("❌ Save cancelled")
			return nil
		}
		// Delete existing template
		os.RemoveAll(templatePath)
	}

	os.MkdirAll(templatePath, 0755)

	// Save template metadata
	metadata := map[string]interface{}{
		"name":        name,
		"created":     time.Now().Format(time.RFC3339),
		"source":      sourcePath,
		"description": "Custom template",
	}
	metadataJSON, _ := yaml.Marshal(metadata)
	os.WriteFile(filepath.Join(templatePath, "metadata.yaml"), metadataJSON, 0644)

	// Copy structure (simplified version)
	err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(sourcePath, path)
		destPath := filepath.Join(templatePath, relPath)
		if d.IsDir() {
			os.MkdirAll(destPath, 0755)
		} else {
			data, _ := os.ReadFile(path)
			os.WriteFile(destPath, data, 0644)
		}
		return nil
	})
	return err
}
func deleteTemplate(name string) {
	homeDir, _ := os.UserHomeDir()
	templatePath := filepath.Join(homeDir, ".omarchy", "templates", name)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Printf("❌ Template '%s' not found\n", name)
		return
	}

	// Confirm deletion
	fmt.Printf("⚠️ Are you sure you want to delete template '%s'? (y/N): ", name)
	var resp string
	fmt.Scanln(&resp)
	if resp != "y" && resp != "Y" {
		fmt.Println("❌ Deletion cancelled")
		return
	}

	// Delete template directory
	err := os.RemoveAll(templatePath)
	if err != nil {
		fmt.Printf("❌ Failed to delete template: %v\n", err)
		return
	}

	fmt.Printf("✅ Template '%s' deleted successfully\n", name)
}
func listTemplates() {
	homeDir, _ := os.UserHomeDir()
	templatesDir := filepath.Join(homeDir, ".omarchy", "templates")

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		fmt.Println("📁 No templates saved yet")
		fmt.Println("   Run: omarchy save <name> to save current project as template")
		return
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		fmt.Printf("❌ Failed to list templates: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("📁 No templates saved yet")
		return
	}

	fmt.Println("📁 Saved Templates:")
	for _, entry := range entries {
		if entry.IsDir() {
			// Try to read metadata
			metadataPath := filepath.Join(templatesDir, entry.Name(), "metadata.yaml")
			if data, err := os.ReadFile(metadataPath); err == nil {
				var metadata map[string]interface{}
				yaml.Unmarshal(data, &metadata)
				if created, ok := metadata["created"]; ok {
					fmt.Printf("  📂 %s (saved: %s)\n", entry.Name(), created)
					continue
				}
			}
			fmt.Printf("  📂 %s\n", entry.Name())
		}
	}
}
func RunGoBuild(name string) error {
	outputName := name

	// If no name provided (empty), try to get from go.mod
	if outputName == "" {
		// Read go.mod to get module name
		data, err := os.ReadFile("go.mod")
		if err == nil {
			lines := strings.Split(string(data), "\n")
			if len(lines) > 0 && strings.HasPrefix(lines[0], "module ") {
				outputName = strings.TrimSpace(strings.TrimPrefix(lines[0], "module "))
			}
		}
	}

	if outputName == "" {
		outputName = "app"
	}

	// Build the binary
	buildCmd := exec.Command("go", "build", "-o", outputName)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		printError("Build failed. Are you in a Go module directory?")
		printInfo("Run: go mod init <module-name> first")
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("✅ Built binary: %s\n", outputName)

	// Install to GOPATH/bin
	installCmd := exec.Command("go", "install")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		printWarning("Install failed, but binary was built")
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

  Git:
    omarchy sync [-a] [-m "message"]             Auto commit and push changes
    omarchy sync --dry-run                       Preview what would happen

  Utilities:
    omarchy doctor                               Check development environment
    omarchy count [ext] [-r]                     Count files by extension
    omarchy tree                                 Show directory tree
    omarchy version                              Show version

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

CONFIGURATION:
  ~/.omarchy.yaml      Default settings (author, license, default_type, etc.)

VERSION:
  Omarchy %s

For more information: https://github.com/Taha95-dev/Omarchy-CLI-tool
`, Version)
}
