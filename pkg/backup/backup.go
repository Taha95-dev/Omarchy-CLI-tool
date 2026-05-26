package backup

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BackupOptions struct {
	Source      string
	Dest        string
	Name        string
	ExcludeDirs []string
}

// Simple progress counter
type progressCounter struct {
	total       int64
	current     int64
	lastPercent int
	lastMsg     string
}

func (p *progressCounter) update(n int64) {
	p.current += n
	if p.total > 0 {
		percent := int(float64(p.current) / float64(p.total) * 100)
		if percent > p.lastPercent && percent < 100 {
			fmt.Printf("\r   Progress: %d%% (%d/%d MB)", percent,
				p.current/(1024*1024), p.total/(1024*1024))
			p.lastPercent = percent
		}
	}
}

func (p *progressCounter) finish() {
	fmt.Printf("\r   Progress: 100%% (%d MB)        \n", p.total/(1024*1024))
}
func CreateBackup(opts BackupOptions) error {
	// Set defaults
	if opts.Source == "" {
		opts.Source = "."
	}

	// Get absolute source path
	absSource, err := filepath.Abs(opts.Source)
	if err != nil {
		return fmt.Errorf("failed to get source path: %w", err)
	}

	// Set destination
	if opts.Dest == "" {
		opts.Dest = "."
	}

	// Generate backup name
	if opts.Name == "" {
		baseName := filepath.Base(absSource)
		timestamp := time.Now().Format("2006-01-02_150405")
		opts.Name = fmt.Sprintf("%s-backup-%s.zip", baseName, timestamp)
	}

	// Ensure .zip extension
	if !strings.HasSuffix(opts.Name, ".zip") {
		opts.Name += ".zip"
	}

	// Full output path
	outputPath := filepath.Join(opts.Dest, opts.Name)

	// Default exclude directories
	if len(opts.ExcludeDirs) == 0 {
		opts.ExcludeDirs = []string{
			"node_modules", ".git", ".vs", "tmp", "cache",
			"__pycache__", ".venv", "venv", "vendor",
		}
	}

	fmt.Printf("📦 Creating backup...\n")
	fmt.Printf("   Source: %s\n", absSource)
	fmt.Printf("   Destination: %s\n", outputPath)

	var totalSize int64
	var totalFiles int64
	filepath.Walk(absSource, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			for _, exclude := range opts.ExcludeDirs {
				if info.Name() == exclude {
					return filepath.SkipDir
				}
			}
			return nil
		}
		totalSize += info.Size()
		totalFiles++
		return nil
	})

	fmt.Printf("   Total: %d files, %.1f MB\n", totalFiles, float64(totalSize)/(1024*1024))

	// Create zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	progress := &progressCounter{total: totalSize}
	var processed int64

	err = filepath.Walk(absSource, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			for _, exclude := range opts.ExcludeDirs {
				if info.Name() == exclude {
					fmt.Printf("   Skipping: %s/\n", exclude)
					return filepath.SkipDir
				}
			}
			return nil
		}

		relPath, err := filepath.Rel(absSource, path)
		if err != nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return nil
		}

		written, err := io.Copy(zipEntry, file)
		if err != nil {
			return nil
		}
		processed += written
		progress.update(processed)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Close zip writer explicitly to flush all data
	zipWriter.Close()

	// Add a small delay for file system sync
	time.Sleep(100 * time.Millisecond)

	// Get final size
	stat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to get backup size: %w", err)
	}
	finalSizeMB := float64(stat.Size()) / (1024 * 1024)

	fmt.Printf("\n✅ Backup created successfully!\n")
	fmt.Printf("   File: %s\n", opts.Name)
	fmt.Printf("   Size: %.2f MB\n", finalSizeMB)
	fmt.Printf("   Files backed up: %d\n", totalFiles)
	fmt.Printf("   Location: %s\n", opts.Dest)

	return nil
}
