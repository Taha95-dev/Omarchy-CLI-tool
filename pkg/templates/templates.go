package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func SaveTemplate(name string, sourcePath string) error {
	// Create templates directory
	homeDir, _ := os.UserHomeDir()
	templatesDir := filepath.Join(homeDir, ".omarchy", "templates")
	os.MkdirAll(templatesDir, 0755)

	// Create template folder
	templatePath := filepath.Join(templatesDir, name)
	if _, err := os.Stat(templatePath); err == nil {
		fmt.Printf("⚠️ Template '%s' already exists. Overwrite? (y/N): ", name)
		var resp string
		fmt.Scanln(&resp)
		if resp != "y" && resp != "Y" {
			fmt.Println("❌ Save cancelled")
			return nil
		}
		// Delete existing template
		os.RemoveAll(templatePath)
	}

	os.MkdirAll(templatePath, 0755)

	// Save template metadata
	metadata := map[string]interface{}{
		"name":        name,
		"created":     time.Now().Format(time.RFC3339),
		"source":      sourcePath,
		"description": "Custom template",
	}
	metadataJSON, _ := yaml.Marshal(metadata)
	os.WriteFile(filepath.Join(templatePath, "metadata.yaml"), metadataJSON, 0644)

	// Copy structure (simplified version)
	err := filepath.WalkDir(sourcePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(sourcePath, path)
		destPath := filepath.Join(templatePath, relPath)
		if d.IsDir() {
			os.MkdirAll(destPath, 0755)
		} else {
			data, _ := os.ReadFile(path)
			os.WriteFile(destPath, data, 0644)
		}
		return nil
	})
	return err
}
func DeleteTemplate(name string) {
	homeDir, _ := os.UserHomeDir()
	templatePath := filepath.Join(homeDir, ".omarchy", "templates", name)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		fmt.Printf("❌ Template '%s' not found\n", name)
		return
	}

	// Confirm deletion
	fmt.Printf("⚠️ Are you sure you want to delete template '%s'? (y/N): ", name)
	var resp string
	fmt.Scanln(&resp)
	if resp != "y" && resp != "Y" {
		fmt.Println("❌ Deletion cancelled")
		return
	}

	// Delete template directory
	err := os.RemoveAll(templatePath)
	if err != nil {
		fmt.Printf("❌ Failed to delete template: %v\n", err)
		return
	}

	fmt.Printf("✅ Template '%s' deleted successfully\n", name)
}
func ListTemplates() {
	homeDir, _ := os.UserHomeDir()
	templatesDir := filepath.Join(homeDir, ".omarchy", "templates")

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		fmt.Println("📁 No templates saved yet")
		fmt.Println("   Run: omarchy save <name> to save current project as template")
		return
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		fmt.Printf("❌ Failed to list templates: %v\n", err)
		return
	}

	if len(entries) == 0 {
		fmt.Println("📁 No templates saved yet")
		return
	}

	fmt.Println("📁 Saved Templates:")
	for _, entry := range entries {
		if entry.IsDir() {
			// Try to read metadata
			metadataPath := filepath.Join(templatesDir, entry.Name(), "metadata.yaml")
			if data, err := os.ReadFile(metadataPath); err == nil {
				var metadata map[string]interface{}
				yaml.Unmarshal(data, &metadata)
				if created, ok := metadata["created"]; ok {
					fmt.Printf("  📂 %s (saved: %s)\n", entry.Name(), created)
					continue
				}
			}
			fmt.Printf("  📂 %s\n", entry.Name())
		}
	}
}
func CreateFromTemplate(templateName, projectName string) error {
    homeDir, _ := os.UserHomeDir()
    templatePath := filepath.Join(homeDir, ".omarchy", "templates", templateName)

    // Check if template exists
    if _, err := os.Stat(templatePath); os.IsNotExist(err) {
        return fmt.Errorf("template '%s' not found", templateName)
    }

    // Create project directory in current location
    projectPath := filepath.Join(".", projectName)
    if _, err := os.Stat(projectPath); err == nil {
        return fmt.Errorf("directory '%s' already exists", projectName)
    }

    // Copy template to new project
    err := filepath.WalkDir(templatePath, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        relPath, _ := filepath.Rel(templatePath, path)
        destPath := filepath.Join(projectPath, relPath)

        if d.IsDir() {
            os.MkdirAll(destPath, 0755)
        } else {
            data, err := os.ReadFile(path)
            if err != nil {
                return err
            }
            os.WriteFile(destPath, data, 0644)
        }
        return nil
    })

    if err != nil {
        return fmt.Errorf("failed to copy template: %w", err)
    }

    // Remove metadata.yaml from copied project (optional)
    metadataPath := filepath.Join(projectPath, "metadata.yaml")
    os.Remove(metadataPath)

    return nil
}
