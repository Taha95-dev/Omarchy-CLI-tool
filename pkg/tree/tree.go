package tree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Print(ctx context.Context, root string, maxDepth int) {
	PrintTree(ctx, root, "", 0, maxDepth)
}

func PrintTree(ctx context.Context, path string, prefix string, currentDepth int, maxDepth int) {
	select {
	case <-ctx.Done():
		fmt.Println("\n❌ Operation cancelled")
		return
	default:
	}

	// Stop if max depth reached
	if maxDepth > 0 && currentDepth >= maxDepth {
		fmt.Printf("%s└── ... (max depth %d reached)\n", prefix, maxDepth)
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	for i, entry := range entries {
		// Skip hidden files/directories (starts with .)
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Also skip specific large folders
		if entry.Name() == "node_modules" || entry.Name() == ".git" || entry.Name() == ".vs" {
			fmt.Printf("%s├── [%s/ (skipped)]\n", prefix, entry.Name())
			continue
		}

		isLast := i == len(entries)-1

		if isLast {
			fmt.Printf("%s└── %s\n", prefix, entry.Name())
			if entry.IsDir() {
				newPrefix := prefix + "    "
				PrintTree(ctx, filepath.Join(path, entry.Name()), newPrefix, currentDepth+1, maxDepth)
			}
		} else {
			fmt.Printf("%s├── %s\n", prefix, entry.Name())
			if entry.IsDir() {
				newPrefix := prefix + "│   "
				PrintTree(ctx, filepath.Join(path, entry.Name()), newPrefix, currentDepth+1, maxDepth)
			}
		}
	}
}

type TreeNode struct {
	Name     string
	IsDir    bool
	Children []*TreeNode
	Indent   int
}

func ParseTree(lines []string) (*TreeNode, error) {
	root := &TreeNode{Name: ".", IsDir: true}
	stack := []*TreeNode{root}

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Calculate depth from prefix (│, ├, └, spaces)
		depth := CalculateDepth(line)

		// Extract name (remove tree symbols)
		name := ExtractName(line)
		isDir := strings.HasSuffix(name, "/")
		if isDir {
			name = strings.TrimSuffix(name, "/")
		}

		node := &TreeNode{Name: name, IsDir: isDir}

		// Adjust stack to correct depth
		for len(stack) > depth+1 {
			stack = stack[:len(stack)-1]
		}

		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)

		if isDir {
			stack = append(stack, node)
		}
	}
	return root, nil
}
func CalculateDepth(line string) int {
	depth := 0
	i := 0
	for i < len(line) {
		if strings.HasPrefix(line[i:], "│   ") {
			depth++
			i += 4
		} else if strings.HasPrefix(line[i:], "    ") {
			depth++
			i += 4
		} else if strings.HasPrefix(line[i:], "├── ") ||
			strings.HasPrefix(line[i:], "└── ") {
			break
		} else {
			break
		}
	}
	return depth
}

func ExtractName(line string) string {
	result := line

	// Clean characters step-by-step from the left side only
	for {
		oldLen := len(result)
		result = strings.TrimPrefix(result, "│   ")
		result = strings.TrimPrefix(result, "    ")
		result = strings.TrimPrefix(result, "├── ")
		result = strings.TrimPrefix(result, "└── ")

		// If the string length stopped changing, we stripped the whole prefix!
		if len(result) == oldLen {
			break
		}
	}
	return strings.TrimSpace(result)
}
func BuildFromTree(root *TreeNode, basePath string) error {
	for _, child := range root.Children {
		fullPath := filepath.Join(basePath, child.Name)

		if child.IsDir {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", child.Name, err)
			}
			fmt.Printf("📁 Created: %s\n", child.Name)

			if err := BuildFromTree(child, fullPath); err != nil {
				return err
			}
		} else {
			// Create empty file
			f, err := os.Create(fullPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", child.Name, err)
			}
			f.Close()
			fmt.Printf("📄 Created: %s\n", child.Name)
		}
	}
	return nil
}

// PrintTreePreview shows the parsed tree structure without creating files
func PrintTreePreview(node *TreeNode, indent string) {
	for i, child := range node.Children {
		isLast := i == len(node.Children)-1

		if isLast {
			fmt.Printf("%s└── %s", indent, child.Name)
		} else {
			fmt.Printf("%s├── %s", indent, child.Name)
		}

		if child.IsDir {
			fmt.Print("/")
		}
		fmt.Println()

		newIndent := indent
		if isLast {
			newIndent += "    "
		} else {
			newIndent += "│   "
		}

		PrintTreePreview(child, newIndent)
	}
}
