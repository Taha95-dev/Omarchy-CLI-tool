package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	config := loadConfig()

	// Set defaults: config value, or fallback to hardcoded
	defaultType := config.DefaultType
	if defaultType == "" {
		defaultType = "cli"
	}

	// Define flags with config defaults
	projectType := flag.String("type", defaultType, "Project type: web, cli, lib, fullstack")
	// Define command line flags
	projectName := flag.String("name", "my-project", "Project name")
	withGit := flag.Bool("git", false, "Initialize git repository")
	withReact := flag.Bool("react", false, "Add React frontend")
	withGo := flag.Bool("go", false, "Add Go backend")

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
		createFullstackStructure(rootPath, *withReact, *withGo)
	default:
		createDefaultStructure(rootPath)
	}

	// Initialize git if requested
	if *withGit {
		initGit(rootPath)
	}

	fmt.Printf("✅ Created %s project: %s\n", *projectType, *projectName)
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

func createFullstackStructure(path string, withReact, withGo bool) {
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
	DefaultGo    bool   `yaml:"default_go"`
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
