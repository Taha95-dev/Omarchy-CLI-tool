package disk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DirInfo struct {
	Path  string
	Size  int64
	Depth int
}

func FormatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	switch exp {
	case 0:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(div))
	case 1:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(div))
	case 2:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(div))
	case 3:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(div))
	}
	return fmt.Sprintf("%d B", bytes)
}

// getDirSize uses PowerShell to get directory size (Windows only)
func getDirSize(path string) int64 {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		cleanPath = path
	}

	// Use PowerShell to get total size
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("(Get-ChildItem -Path '%s' -Recurse -File | Measure-Object -Property Length -Sum).Sum", cleanPath))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	sizeStr := strings.TrimSpace(string(output))
	if sizeStr == "" {
		return 0
	}

	var size int64
	fmt.Sscanf(sizeStr, "%d", &size)
	return size
}

// getDirSizeSimple uses filepath.Walk (falls back if PowerShell fails)
func getDirSizeSimple(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func printTree(path string, maxDepth int, currentDepth int, prefix string, isLast bool) {
	if maxDepth > 0 && currentDepth > maxDepth {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		return
	}

	// Use both methods to get size
	size := getDirSize(path)
	if size == 0 {
		size = getDirSizeSimple(path)
	}

	base := filepath.Base(path)

	// Skip hidden files/directories
	if strings.HasPrefix(base, ".") && base != "." && base != ".." {
		return
	}

	// Skip large system folders
	skipDirs := []string{"node_modules", ".git", ".vs", "Library", "AppData", "System32", "Windows"}
	for _, skip := range skipDirs {
		if base == skip {
			fmt.Printf("%s├── [%s/ (skipped)]\n", prefix, base)
			return
		}
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if info.IsDir() {
		if size == 0 {
			fmt.Printf("%s%s%s/ (empty or inaccessible)\n", prefix, connector, base)
		} else {
			fmt.Printf("%s%s%s/ (%s)\n", prefix, connector, base, FormatSize(size))
		}

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}

		for i, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			isLastChild := i == len(entries)-1
			printTree(childPath, maxDepth, currentDepth+1, newPrefix, isLastChild)
		}
	} else {
		fmt.Printf("%s%s%s (%s)\n", prefix, connector, base, FormatSize(size))
	}
}

func ShowDiskUsage(path string, depth int, human bool) error {
	cleanPath, err := filepath.Abs(path)
	if err != nil {
		cleanPath = path
	}

	info, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Get total size using PowerShell
	totalSize := getDirSize(cleanPath)
	if totalSize == 0 {
		// Fallback to filepath.Walk
		totalSize = getDirSizeSimple(cleanPath)
	}

	fmt.Printf("📁 Disk Usage for: %s\n", cleanPath)

	if depth > 0 && info.IsDir() {
		fmt.Printf("🔍 Max Depth: %d\n", depth)
		entries, err := os.ReadDir(cleanPath)
		if err != nil {
			return err
		}

		for i, entry := range entries {
			childPath := filepath.Join(cleanPath, entry.Name())
			isLast := i == len(entries)-1
			printTree(childPath, depth, 1, "", isLast)
		}
	}

	fmt.Printf("\n📊 Total: %s\n", FormatSize(totalSize))
	return nil
}
