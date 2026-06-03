package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
)

type CleanupOptions struct {
	DryRun bool
	All    bool
}

type DeletedItem struct {
	Path   string
	Size   int64
	Reason string
}

func RunCleanup(path string, opts CleanupOptions) ([]DeletedItem, int64, error) {
	var deleted []DeletedItem
	var totalSize int64

	// Simple lookup map for O(1) folder checking speed
	isTargetFolder := map[string]bool{
		"node_modules": true,
		".cache":       true,
		"dist":         true,
		"build":        true,
		"__pycache__":  true,
	}

	// Run a full structural recursive walk down the workspace directory tree
	err := filepath.WalkDir(path, func(currentPath string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip items we can't read rather than breaking the whole tool
		}

		// Prevent scanning your own root folder
		if currentPath == path {
			return nil
		}

		base := d.Name()

		// 1. Check for Heavy Artifact Directories if --all is active
		if d.IsDir() {
			if opts.All && isTargetFolder[base] {
				// Safeguard check
				if base == ".git" {
					if !confirmDelete(currentPath, "WARNING: This is your git repository!") {
						return filepath.SkipDir
					}
				}

				info, err := d.Info()
				var size int64
				if err == nil {
					size = info.Size() // Note: Directory size might only read block sizes
				}

				if opts.DryRun {
					fmt.Printf("   Would delete directory: %s\n", currentPath)
				} else {
					if err := os.RemoveAll(currentPath); err != nil {
						fmt.Printf("   ❌ Failed to delete folder: %s (%v)\n", currentPath, err)
						return filepath.SkipDir
					}
					fmt.Printf("   🗑️ Deleted directory: %s\n", currentPath)
				}

				deleted = append(deleted, DeletedItem{
					Path:   currentPath,
					Size:   size,
					Reason: "directory cleanup",
				})
				totalSize += size

				return filepath.SkipDir // We deleted the folder! Don't look inside it.
			}
			return nil // Keep walking deeper into non-matching folders
		}

		// 2. File Level Target matching (*.tmp, *.log, etc.)
		shouldDeleteFile := false
		ext := filepath.Ext(base)

		if ext == ".tmp" || ext == ".temp" || ext == ".cache" || ext == ".log" || ext == ".old" || ext == ".backup" {
			shouldDeleteFile = true
		}

		if shouldDeleteFile {
			info, err := d.Info()
			if err != nil {
				return nil
			}

			if opts.DryRun {
				fmt.Printf("   Would delete file: %s (%s)\n", currentPath, FormatSize(info.Size()))
			} else {
				if err := os.Remove(currentPath); err != nil {
					fmt.Printf("   ❌ Failed to delete file: %s (%v)\n", currentPath, err)
					return nil
				}
				fmt.Printf("   🗑️ Deleted file: %s (%s)\n", currentPath, FormatSize(info.Size()))
			}

			deleted = append(deleted, DeletedItem{
				Path:   currentPath,
				Size:   info.Size(),
				Reason: "file cleanup",
			})
			totalSize += info.Size()
		}

		return nil
	})

	return deleted, totalSize, err
}

func FormatSize(bytes int64) string {
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

func confirmDelete(path string, warning string) bool {
	fmt.Printf("   ⚠️  %s: %s\n", warning, path)
	fmt.Print("      Delete anyway? (y/N): ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}
