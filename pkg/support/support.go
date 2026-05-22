package support

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func PrintError(msg string) {
	fmt.Println("❌", msg)
}

func PrintWarning(msg string) {
	fmt.Println("⚠️", msg)
}

func PrintSuccess(msg string) {
	fmt.Println("✅", msg)
}
func PrintSuccessf(format string, args ...interface{}) {
	fmt.Printf("✅ "+format+"\n", args...)
}
func PrintInfo(msg string) {
	fmt.Println("ℹ️", msg)
}
func RunCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("⚠️ Warning:", err)
	}
}
func GetHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".omarchy.yaml")
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
func GetOS() string {
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
