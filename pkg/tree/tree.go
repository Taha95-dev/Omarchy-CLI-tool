package tree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Print(ctx context.Context, root string) {
	PrintTree(ctx, root, "")
}

func PrintTree(ctx context.Context, path string, prefix string) {
	select {
	case <-ctx.Done():
		fmt.Println("\n❌ Operation cancelled")
		return
	default:
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("❌ Error reading directory: %v\n", err)
		return
	}

	for i, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		isLast := i == len(entries)-1

		if isLast {
			fmt.Printf("%s└── %s\n", prefix, entry.Name())
			newPrefix := prefix + "    "
			if entry.IsDir() {
				PrintTree(ctx, filepath.Join(path, entry.Name()), newPrefix)
			}
		} else {
			fmt.Printf("%s├── %s\n", prefix, entry.Name())
			newPrefix := prefix + "│   "
			if entry.IsDir() {
				PrintTree(ctx, filepath.Join(path, entry.Name()), newPrefix)
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
	// Remove prefix symbols
	patterns := []string{"├── ", "└── ", "│   ", "    "}
	result := line
	for _, p := range patterns {
		result = strings.ReplaceAll(result, p, "")
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
