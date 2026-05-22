package counter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

func Count(suffix string) (int, error) {
	suffix = strings.TrimPrefix(suffix, ".")
	entries, err := os.ReadDir(".")
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "."+suffix) {
			count++
		}
	}
	return count, nil
}

func CountRecursive(ctx context.Context, suffix string) (int, error) {
	suffix = strings.TrimPrefix(suffix, ".")
	count := 0

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err != nil {
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), "."+suffix) {
				count++
			}
			return nil
		}
	})

	return count, err
}
