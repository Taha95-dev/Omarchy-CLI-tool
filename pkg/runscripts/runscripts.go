package runscripts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var scripts = map[string]map[string]string{
	"node":   {"dev": "npm run dev", "build": "npm run build", "test": "npm test", "start": "npm start"},
	"go":     {"dev": "go run .", "build": "go build -o app", "test": "go test ./...", "start": "./app"},
	"python": {"dev": "python app.py", "test": "pytest", "start": "python app.py"},
	"cpp":    {"dev": "g++ -std=c++23 main.cpp -o main.exe && main.exe", "build": "g++ -std=c++23 main.cpp -o main.exe", "start": "main.exe"},
	"csharp": {"dev": "dotnet run", "build": "dotnet build", "start": "dotnet run"},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// DetectProjectType returns the project type based on config files
func DetectProjectType() string {
	switch {
	case fileExists("package.json"):
		fmt.Println("📦 Detected Node.js project")
		return "node"
	case fileExists("go.mod"):
		fmt.Println("📦 Detected Go project")
		return "go"
	case fileExists("requirements.txt"):
		fmt.Println("📦 Detected Python project")
		return "python"
	case fileExists("CMakeLists.txt") || fileExists("main.cpp"):
		fmt.Println("📦 Detected C++ project")
		return "cpp"
	case hasExtension(".csproj"):
		fmt.Println("📦 Detected C# project")
		return "csharp"
	default:
		fmt.Println("❌ Unknown project type. No matching configurations found.")
		return "unknown"
	}
}

// Helper to look for wildcard extensions like *.csproj
func hasExtension(ext string) bool {
	files, err := os.ReadDir(".")
	if err != nil {
		return false
	}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ext) {
			return true
		}
	}
	return false
}

// ListScripts prints all available scripts for the detected project type
func ListScripts(projectType string) {
	cmds, ok := scripts[projectType]
	if !ok {
		fmt.Printf("❌ Unknown project type: %s\n", projectType)
		return
	}

	fmt.Println("\n📋 Available scripts:")
	for name, cmd := range cmds {
		fmt.Printf("  %-6s → %s\n", name, cmd)
	}
}

// RunScript executes a specific script by name through the system shell
func RunScript(projectType, scriptName string) error {
	cmds, ok := scripts[projectType]
	if !ok {
		fmt.Printf("❌ Unknown project type: %s\n", projectType)
		return fmt.Errorf("unknown project type: %s", projectType)
	}

	cmdStr, ok := cmds[scriptName]
	if !ok {
		fmt.Printf("❌ Unknown script: %s\nAvailable: dev, build, test, start\n", scriptName)
		return fmt.Errorf("unknown script: %s", scriptName)
	}

	fmt.Printf("🚀 Running: %s\n", cmdStr)

	var cmd *exec.Cmd
	// Windows uses 'cmd /C', Mac/Linux uses 'sh -c'
	if os.PathSeparator == '\\' {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ Failed to run script: %v\n", err)
		return fmt.Errorf("failed to run script: %w", err)
	}

	fmt.Println("✅ Done")
	return nil
}
