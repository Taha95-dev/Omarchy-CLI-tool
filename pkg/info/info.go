package info

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ShowInfo displays project statistics
func ShowInfo() {
	// Count files (excluding .git, node_modules)
	files := countFilesExcluding(".", []string{".git", "node_modules", "vendor"})

	// Count lines of code
	lines := countLinesInFilesExcluding(".", []string{".git", "node_modules", "vendor"})

	// Find TODOs
	todos := findTodosExcluding(".", []string{".git", "node_modules", "vendor"})

	// Git info
	branch := getCurrentBranch()
	lastCommit := getLastCommitTime()

	// Print
	fmt.Printf("\n📊 Project Info\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("📁 Files:         %d\n", files)
	fmt.Printf("🧮 Lines:         %d\n", lines)
	fmt.Printf("🐛 TODOs:         %d\n", todos)
	fmt.Printf("🌿 Branch:        %s\n", branch)
	fmt.Printf("⏰ Last commit:   %s\n", lastCommit)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func countFilesExcluding(root string, excludeDirs []string) int {
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Check if this directory should be excluded
		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if info.Name() == exclude {
					return filepath.SkipDir
				}
			}
			return nil
		}

		count++
		return nil
	})
	if err != nil {
		fmt.Printf("⚠️ Error counting files: %v\n", err)
	}
	return count
}

func countLinesInFilesExcluding(root string, excludeDirs []string) int {
	lines := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if info.Name() == exclude {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only count code files (skip binaries, images, etc.)
		if isCodeFile(info.Name()) {
			lines += countLinesInFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("⚠️ Error counting lines: %v\n", err)
	}
	return lines
}

func isCodeFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	codeExts := []string{".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".rs", ".md", ".txt", ".json", ".yaml", ".yml", ".html", ".css", ".scss"}
	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

func countLinesInFile(filename string) int {
	lines := 0
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines++
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("⚠️ Error reading %s: %v\n", filename, err)
		return 0
	}

	return lines
}

func findTodosInFile(filename string) int {
	count := 0
	file, err := os.Open(filename)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for TODO, FIXME, HACK
		if strings.Contains(strings.ToUpper(line), "TODO") ||
			strings.Contains(strings.ToUpper(line), "FIXME") ||
			strings.Contains(strings.ToUpper(line), "HACK") {
			count++
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("⚠️ Error scanning %s for TODOs: %v\n", filename, err)
		return 0
	}

	return count
}

func findTodosExcluding(root string, excludeDirs []string) int {
	todos := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			for _, exclude := range excludeDirs {
				if info.Name() == exclude {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if isCodeFile(info.Name()) {
			todos += findTodosInFile(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("⚠️ Error finding TODOs: %v\n", err)
	}
	return todos
}

func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "not a git repo"
	}
	return strings.TrimSpace(string(output))
}

func getLastCommitTime() string {
	cmd := exec.Command("git", "log", "-1", "--format=%cd", "--date=relative")
	output, err := cmd.Output()
	if err != nil {
		return "no commits"
	}
	return strings.TrimSpace(string(output))
}
