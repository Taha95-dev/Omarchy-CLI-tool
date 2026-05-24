package doctor

import (
	"context"
	"fmt"
	"omarchy/pkg/gitsupport"
	"omarchy/pkg/support"
	"os"
)

func RunDoctor(ctx context.Context) {
	select {
	case <-ctx.Done():
		fmt.Println("❌ Doctor check cancelled")
		return
	default:
	}

	fmt.Printf("🖥️ Operating System: %s\n", support.GetOS())
	fmt.Printf("🏠 Home Directory: %s\n", support.GetHomeDir())
	fmt.Printf("🔍 Omarchy Environment Check\n")

	// Tool checks
	support.CheckTool("node", "--version")
	support.CheckTool("go", "version")
	support.CheckTool("git", "--version")
	support.CheckTool("npm", "--version")
	support.CheckTool("docker", "--version")

	// Config check
	configPath := support.GetConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("✅ Omarchy config found at: %s\n", configPath)
	} else {
		fmt.Printf("⚠️ Omarchy config not found (run: omarchy config init)\n")
	}

	// Git config check
	fmt.Printf("📋 Git Configuration:\n")
	gitsupport.CheckGitConfig()

	// === NEW SAFETY CHECKS ===

	// 1. Git in home directory
	fmt.Printf("🏠 Home Directory Safety:\n")
	gitsupport.CheckGitInHome()

	// 2. Git in common wrong places
	fmt.Printf("📁 Project Location Safety:\n")
	gitsupport.CheckGitInDesktop()
	gitsupport.CheckGitInDownloads()
	gitsupport.CheckGitInDocuments()

	// 3. HUGE node_modules
	fmt.Printf("📦 Dependency Health:\n")
	gitsupport.CheckNodeModulesSize()

	// 4. Disk space
	fmt.Printf("💾 Disk Space:\n")
	gitsupport.CheckDiskSpace()

	// 5. Environment variables
	fmt.Printf("🌍 Environment Variables:\n")
	gitsupport.CheckCommonEnvVars()

	// 6. Large files in git
	fmt.Printf("📄 Git Repository Health:\n	")
	gitsupport.CheckLargeFilesInGit()
	gitsupport.CheckUntrackedFiles()

	// 7. VS Code extensions
	fmt.Printf("🧩 VS Code Extensions:\n")
	gitsupport.CheckVSCodeExtensions()

	// 8. Omarchy version
	fmt.Printf("🚀 Omarchy Health:\n")
	gitsupport.CheckOmarchyVersion()
}
