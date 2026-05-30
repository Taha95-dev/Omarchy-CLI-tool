package deploy

import (
	"fmt"
	"omarchy/pkg/database"
	"os"
	"os/exec"
	"strings"
)

type Platform string

const (
	Render  Platform = "render"
	Netlify Platform = "netlify"
	Vercel  Platform = "vercel"
	Github  Platform = "github"
)

type ProjectType string

const (
	ReactApp   ProjectType = "react-app"
	NextJsApp  ProjectType = "nextjs-app"
	GoApp      ProjectType = "go-app"
	NodeApp    ProjectType = "node-app"
	StaticSite ProjectType = "static-site"
	Unknown    ProjectType = "unknown"
)

type DeployConfig struct {
	Platform     Platform
	ProjectType  ProjectType
	ProjectPath  string
	EnvVars      map[string]string
	BuildCommand string
	OutputDir    string
}

func DetectProjectType() ProjectType {
	if _, err := os.Stat("package.json"); err == nil {
		data, _ := os.ReadFile("package.json")
		content := string(data)
		if strings.Contains(content, "\"react\"") {
			return ReactApp
		}
		if strings.Contains(content, "\"next\"") {
			return NextJsApp
		}
		return NodeApp
	}
	// Check for go
	if _, err := os.Stat("go.mod"); err == nil {
		return GoApp
	}
	// Check for static site
	if _, err := os.Stat("index.html"); err == nil {
		return StaticSite
	}
	return Unknown
}

func DetectPlatform() Platform {
	// Check for existing configs
	if _, err := os.Stat("render.yaml"); err == nil {
		return Render
	}
	if _, err := os.Stat("netlify.toml"); err == nil {
		return Netlify
	}
	// Default Render Free Tier
	return Render
}

func GenerateConfig(projectType ProjectType, platform Platform) error {
	switch platform {
	case Render:
		return GenerateRenderYaml(projectType)
	case Netlify:
		return GenerateNetlifyToml(projectType)
	case Vercel:
		return GenerateVercelJson(projectType)
	case Github:
		return GenerateGithubWorkflow(projectType)
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
}

func GenerateRenderYaml(projectType ProjectType) error {
	content := `services:
  - type: web
    name: omarchy-deploy
    runtime:`

	var buildCmd, startCmd, envVars string

	switch projectType {
	case ReactApp, NextJsApp:
		content += ` node
    buildCommand: npm install && npm run build
    startCommand: npm run start
    envVars:
      - key: NODE_VERSION
        value: 18`
	case GoApp:
		content += ` go
    buildCommand: go build -o app
    startCommand: ./app`
	case NodeApp:
		content += ` node
    buildCommand: npm install
    startCommand: npm start`
	default:
		content += ` static
    buildCommand: ""
    startCommand: ""
    staticPublishPath: ./`
	}

	_ = buildCmd
	_ = startCmd
	_ = envVars

	return os.WriteFile("render.yaml", []byte(content), 0644)
}

func GenerateNetlifyToml(projectType ProjectType) error {
	content := `[build]
  publish = "dist"
  command = "npm run build"

[build.environment]
  NODE_VERSION = "18"`

	return os.WriteFile("netlify.toml", []byte(content), 0644)
}

func GenerateVercelJson(projectType ProjectType) error {
	content := `{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "devCommand": "npm run dev",
  "installCommand": "npm install"
}`

	return os.WriteFile("vercel.json", []byte(content), 0644)
}

func GenerateGithubWorkflow(projectType ProjectType) error {
	// Create .github/workflows/deploy.yml
	os.MkdirAll(".github/workflows", 0755)

	content := `name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm install
      - run: npm run build
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./dist`

	return os.WriteFile(".github/workflows/deploy.yml", []byte(content), 0644)
}

// Deploy with auto flag (skip interactive prompts)
func Deploy(platform Platform, projectType ProjectType, projectName string, auto bool) error {
	if auto {
		fmt.Printf("🚀 Auto-deploying %s project to %s...\n", projectType, platform)
	} else {
		fmt.Printf("🚀 Deploying %s project to %s...\n", projectType, platform)
	}

	// Check if config already exists (skip generation if auto)
	configExists := false
	switch platform {
	case Render:
		_, err := os.Stat("render.yaml")
		configExists = err == nil
	case Netlify:
		_, err := os.Stat("netlify.toml")
		configExists = err == nil
	case Vercel:
		_, err := os.Stat("vercel.json")
		configExists = err == nil
	case Github:
		_, err := os.Stat(".github/workflows/deploy.yml")
		configExists = err == nil
	}

	if !configExists {
		// Step 1: Generate config file
		if err := GenerateConfig(projectType, platform); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		fmt.Printf("✅ Generated %s config\n", platform)
	} else if !auto {
		fmt.Printf("⚠️  %s config already exists. Skipping generation.\n", platform)
	}

	// Step 2: Commit and push (if git repo) - skip prompts in auto mode
	if _, err := os.Stat(".git"); err == nil {
		if !auto {
			fmt.Print("Commit and push changes? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "yes" {
				fmt.Println("Skipping git push. You'll need to push manually.")
				goto deployInstructions
			}
		}
		runCmd("git", "add", ".")
		runCmd("git", "commit", "-m", "Add deploy config for "+string(platform))
		runCmd("git", "push")
		fmt.Println("✅ Pushed to GitHub")
	}

deployInstructions:
	// Step 3: Platform-specific instructions (shorter in auto mode)
	switch platform {
	case Render:
		if auto {
			fmt.Println("\n✅ Auto-deploy prepared! Run this command to deploy:")
			fmt.Printf("   cd %s && git push\n", projectName)
			fmt.Println("\n   Then go to https://dashboard.render.com and connect your repo")
		} else {
			fmt.Println("\n🌐 To deploy on Render:")
			fmt.Println("   1. Go to https://render.com")
			fmt.Println("   2. Click 'New +' → 'Web Service'")
			fmt.Println("   3. Connect your GitHub repository")
			fmt.Println("   4. Render will auto-detect render.yaml")
			fmt.Println("   5. Click 'Apply'")
		}
	case Netlify:
		if auto {
			fmt.Println("\n✅ Auto-deploy prepared! Run:")
			fmt.Printf("   cd %s && git push\n", projectName)
			fmt.Println("   Then drag folder to https://app.netlify.com/drop")
		} else {
			fmt.Println("\n🌐 To deploy on Netlify:")
			fmt.Println("   1. Go to https://netlify.com")
			fmt.Println("   2. Drag and drop your project folder")
			fmt.Println("   3. Or connect GitHub")
		}
	case Vercel:
		if auto {
			fmt.Println("\n✅ Auto-deploy prepared! Run:")
			fmt.Printf("   cd %s && vercel --prod\n", projectName)
		} else {
			fmt.Println("\n🌐 To deploy on Vercel:")
			fmt.Println("   1. Go to https://vercel.com")
			fmt.Println("   2. Import your GitHub repository")
			fmt.Println("   3. Vercel auto-detects vercel.json")
		}
	case Github:
		fmt.Println("\n✅ GitHub Actions workflow created!")
		if auto {
			fmt.Println("   Push to main to trigger auto-deploy")
		} else {
			fmt.Println("   Push to main to trigger auto-deploy")
		}
	}

	return nil
}
func PreviewReset(dbType database.DatabaseType) {
	fmt.Println("🔍 DRY RUN: Tables that will be dropped:")

	tables, err := database.GetTableNames(dbType) // You'll need to implement this
	if err != nil {
		fmt.Printf("  ⚠️ Could not list tables: %v\n", err)
		return
	}
	for _, table := range tables {
		rowCount := database.GetTableRowCount(dbType, table)
		fmt.Printf("  - %s (%d rows)\n", table, rowCount)
	}

	if len(tables) == 0 {
		fmt.Println("  (no tables found)")
	}
}

func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
