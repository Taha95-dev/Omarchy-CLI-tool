package find

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FindOptions struct {
	Pattern       string
	Type          string // f for file, d for directory
	Name          string // exact name
	Extension     string // file extension
	Size          string // +10M, -1G, etc.
	ExcludeDirs   []string
	MaxDepth      int
	CaseSensitive bool
	UseRegex      bool
}

type Result struct {
	Path  string
	Size  int64
	IsDir bool
}

func Find(root string, opts FindOptions) ([]Result, error) {
	var results []Result

	// Compile regex if needed
	var regex *regexp.Regexp
	if opts.UseRegex {
		var err error
		regex, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
	}

	// Default exclude dirs
	if len(opts.ExcludeDirs) == 0 {
		opts.ExcludeDirs = []string{".git", "node_modules", ".vs", "tmp", "cache"}
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Check depth
		if opts.MaxDepth > 0 {
			rel, _ := filepath.Rel(root, path)
			depth := strings.Count(rel, string(filepath.Separator))
			if depth > opts.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip excluded dirs
		if d.IsDir() {
			for _, exclude := range opts.ExcludeDirs {
				if d.Name() == exclude {
					return filepath.SkipDir
				}
			}
		}

		// Filter by type
		if opts.Type != "" {
			if opts.Type == "f" && d.IsDir() {
				return nil
			}
			if opts.Type == "d" && !d.IsDir() {
				return nil
			}
		}

		name := d.Name()
		if !opts.CaseSensitive {
			name = strings.ToLower(name)
		}

		match := false

		// Match by pattern
		if opts.Pattern != "" {
			if opts.UseRegex {
				if regex.MatchString(name) {
					match = true
				}
			} else {
				pattern := opts.Pattern
				if !opts.CaseSensitive {
					pattern = strings.ToLower(pattern)
				}
				if strings.Contains(name, pattern) {
					match = true
				}
			}
		}

		// Match by exact name
		if opts.Name != "" {
			target := opts.Name
			if !opts.CaseSensitive {
				target = strings.ToLower(target)
			}
			if name == target {
				match = true
			}
		}

		// Match by extension
		if opts.Extension != "" {
			ext := filepath.Ext(name)
			if ext == "."+opts.Extension || ext == opts.Extension {
				match = true
			}
		}

		// Match by size
		if opts.Size != "" && !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size := info.Size()
				if matchSize(size, opts.Size) {
					match = true
				}
			}
		}

		if match {
			info, _ := d.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			results = append(results, Result{
				Path:  path,
				Size:  size,
				IsDir: d.IsDir(),
			})
		}

		return nil
	})

	return results, err
}

func matchSize(size int64, pattern string) bool {
	pattern = strings.ToLower(pattern)

	// Parse size pattern like +10M, -1G, 500K
	var op string
	var num int64
	var unit string

	if strings.HasPrefix(pattern, "+") {
		op = "+"
		pattern = pattern[1:]
	} else if strings.HasPrefix(pattern, "-") {
		op = "-"
		pattern = pattern[1:]
	}

	// Parse number and unit
	for i, c := range pattern {
		if c >= '0' && c <= '9' {
			continue
		}
		numStr := pattern[:i]
		unit = pattern[i:]
		fmt.Sscanf(numStr, "%d", &num)
		break
	}

	// Convert to bytes
	switch unit {
	case "K", "KB":
		num *= 1024
	case "M", "MB":
		num *= 1024 * 1024
	case "G", "GB":
		num *= 1024 * 1024 * 1024
	case "T", "TB":
		num *= 1024 * 1024 * 1024 * 1024
	}

	switch op {
	case "+":
		return size > num
	case "-":
		return size < num
	default:
		// Exact size (allow small variation)
		return size > num*95/100 && size < num*105/100
	}
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
	}
	return fmt.Sprintf("%d B", bytes)
}

func PrintResults(results []Result, verbose bool) {
	if len(results) == 0 {
		fmt.Println("No matches found")
		return
	}

	for _, r := range results {
		if verbose {
			sizeStr := FormatSize(r.Size)
			typeChar := "📄"
			if r.IsDir {
				typeChar = "📁"
			}
			fmt.Printf("%s %s (%s)\n", typeChar, r.Path, sizeStr)
		} else {
			fmt.Println(r.Path)
		}
	}

	fmt.Printf("\n📊 Found %d matches\n", len(results))
}
