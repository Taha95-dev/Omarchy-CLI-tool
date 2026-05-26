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

	// Patterns to delete
	patterns := []string{
		"*.tmp",
		"*.temp",
		"*.cache",
		"*.log",
		"*.old",
		"*.backup",
		"*-backup-*.zip",
	}

	// Directories to clean (if --all)
	if opts.All {
		patterns = append(patterns, "node_modules/", ".cache/", "dist/", "build/", "__pycache__/")
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(path, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			// Skip if it's a directory and we're not doing --all
			if info.IsDir() && !opts.All {
				continue
			}

			// Don't delete dangerous stuff
			base := filepath.Base(match)
			if base == ".git" || base == ".git/" {
				if !confirmDelete(match, "WARNING: This is your git repository!") {
					continue
				}
			}

			if opts.DryRun {
				fmt.Printf("   Would delete: %s (%s)\n", match, FormatSize(info.Size()))
				deleted = append(deleted, DeletedItem{
					Path:   match,
					Size:   info.Size(),
					Reason: "dry run",
				})
			} else {
				err := os.RemoveAll(match)
				if err != nil {
					fmt.Printf("   ❌ Failed to delete: %s (%v)\n", match, err)
					continue
				}
				fmt.Printf("   🗑️ Deleted: %s (%s)\n", match, FormatSize(info.Size()))
				deleted = append(deleted, DeletedItem{
					Path:   match,
					Size:   info.Size(),
					Reason: "deleted",
				})
			}
			totalSize += info.Size()
		}
	}

	return deleted, totalSize, nil
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
