package tree

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Print(ctx context.Context, root string) {
	printTree(ctx, root, "")
}

func printTree(ctx context.Context, path string, prefix string) {
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
				printTree(ctx, filepath.Join(path, entry.Name()), newPrefix)
			}
		} else {
			fmt.Printf("%s├── %s\n", prefix, entry.Name())
			newPrefix := prefix + "│   "
			if entry.IsDir() {
				printTree(ctx, filepath.Join(path, entry.Name()), newPrefix)
			}
		}
	}
}
