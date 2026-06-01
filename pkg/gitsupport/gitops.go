package gitsupport

import (
	"context"
	"fmt"
	"omarchy/pkg/support"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GitSync(ctx context.Context, autoMsg bool, customMsg, tag string) {
	homeDir, _ := os.UserHomeDir()
	currentDir, _ := os.Getwd()

	if currentDir == homeDir {
		support.PrintError("Cannot run 'omarchy sync' in home directory!")
		support.PrintInfo("This would try to commit your entire home folder.")
		support.PrintInfo("Move to a project folder first: cd ~/Documents/vs code/")
		return
	}

	// Check if .git exists in home
	if _, err := os.Stat(filepath.Join(homeDir, ".git")); err == nil {
		support.PrintWarning("Detected .git folder in home directory!")
		support.PrintInfo("Run: rm -rf ~/.git")
	}

	select {
	case <-ctx.Done():
		fmt.Println("❌ Sync cancelled")
		return
	default:
	}

	if !isGitInstalled() {
		support.PrintError("Git is not installed or not in PATH")
		support.PrintInfo("Install Git from: https://git-scm.com")
		return
	}

	// Check 2: In a git repo?
	inRepo, err := isGitRepo()
	if err != nil {
		support.PrintError(err.Error())
		return
	}
	if !inRepo {
		support.PrintError("Not in a git repository")
		support.PrintInfo("Initializing git repository...")
		support.RunCmd("git", "init")
		return
	}

	// Check 3: Any changes?
	if !hasGitChanges(".") {
		support.PrintInfo("No changes to commit")
		return
	}

	// Check 4: Git user configured?
	userName, userEmail, err := CheckGitConfig()
	if err != nil {
		support.PrintError("Git user not configured")
		return
	}
	fmt.Printf("✅ Git user: %s <%s>\n", userName, userEmail)

	// Add changes
	support.RunCmd("git", "add", ".")

	// Create commit message
	message := customMsg
	if message == "" && autoMsg {
		message = fmt.Sprintf("Auto-sync: %s", time.Now().Format("2006-01-02 15:04:05"))
	}
	if message == "" && !autoMsg {
		message = "Sync via Omarchy"
	}

	// Commit
	support.RunCmd("git", "commit", "-m", message)
	fmt.Println("✅ Committed:", message)

	// Check if remote exists before pushing
	if hasRemote() {
		support.RunCmd("git", "push")
		fmt.Println("✅ Pushed to remote")
	} else {
		support.PrintWarning("No remote configured. Commit saved locally only.")
		support.PrintInfo("Run: git remote add origin <url>")
	}
	// After push, handle tag
	if tag != "" {
		fmt.Printf("🏷️ Checking tag availability: %s\n", tag)

		// Check if tag exists locally or remotely
		if err := exec.Command("git", "rev-parse", tag).Run(); err == nil {
			fmt.Printf("⚠️ Tag %s already exists. Skipping tag creation.\n", tag)
			return
		}

		// Create tag using your secure project utility wrapper
		tagMsg := fmt.Sprintf("Release %s", tag)
		if customMsg != "" {
			tagMsg = customMsg
		}

		fmt.Printf("📦 Stamping repository with tag %s...\n", tag)
		support.RunCmd("git", "tag", "-a", tag, "-m", tagMsg)
		fmt.Printf("✅ Created local tag: %s\n", tag)

		// Push tag securely if a remote exists
		if hasRemote() {
			fmt.Printf("🚀 Uploading tag %s to origin remote tracking branch...\n", tag)
			support.RunCmd("git", "push", "origin", tag)
			fmt.Printf("✅ Pushed tag %s to origin\n", tag)
		} else {
			support.PrintWarning("Skipping tag push: No remote configuration origin detected.")
		}
	}
}

func DryRunSync(ctx context.Context) {
	select {
	case <-ctx.Done():
		fmt.Println("❌ Dry run cancelled")
		return
	default:
	}

	fmt.Println("🔍 Dry run - what would happen:")

	if !isGitInstalled() {
		support.PrintError("Git is not installed or not in PATH")
		support.PrintInfo("Install Git from: https://git-scm.com")
		return
	}

	inRepo, err := isGitRepo()
	if err != nil {
		support.PrintError(err.Error())
		return
	}
	if !inRepo {
		support.PrintError("Not in a git repository")
		support.PrintInfo("Run: git init")
		return
	}

	if hasGitChanges(".") {
		support.PrintSuccess("Would add all changes")
		support.PrintSuccess("Would commit with message: Auto-sync: <timestamp>")

		if hasRemote() {
			support.PrintSuccess("Would push to remote")
		} else {
			support.PrintWarning("Would skip push (no remote)")
		}
	} else {
		support.PrintInfo("No changes to commit")
	}
}
func hasRemote() bool {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	err := cmd.Run()
	return err == nil
}
func isGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func isGitRepo() (bool, error) {
	if !isGitInstalled() {
		return false, fmt.Errorf("git is not installed or not in PATH")
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	if err != nil {
		return false, nil // Not a git repo (no error)
	}
	return true, nil
}
func hasGitChanges(path string) bool {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false // Assume no changes if git fails
	}
	return len(output) > 0 // Has changes if output not empty
}
func CheckGitConfig() (string, string, error) {
	nameCmd := exec.Command("git", "config", "--global", "user.name")
	nameOutput, nameErr := nameCmd.Output()

	emailCmd := exec.Command("git", "config", "--global", "user.email")
	emailOutput, emailErr := emailCmd.Output()

	if nameErr != nil || emailErr != nil {
		support.PrintWarning("Git user.name or user.email not set")
		support.PrintInfo("Run: git config --global user.name \"Your Name\"")
		support.PrintInfo("Run: git config --global user.email \"you@example.com\"")
		return "", "", fmt.Errorf("git config not set")
	}

	userName := strings.TrimSpace(string(nameOutput))
	userEmail := strings.TrimSpace(string(emailOutput))
	support.PrintSuccessf("Git user: %s <%s>", userName, userEmail)
	return userName, userEmail, nil
}
func CheckGitInHome() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	gitPath := filepath.Join(homeDir, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		// No .git in home - good!
		return
	}

	if info.IsDir() {
		support.PrintError("❌ CRITICAL: .git folder found in home directory!")
		support.PrintWarning("   Your entire home folder is being tracked by git.")
		support.PrintInfo("   This will slow down your computer and git commands.")
		support.PrintInfo("   Fix: rm -rf ~/.git")
		support.PrintInfo("   Then: cd ~/Documents/vs\\ code/omarchy")
		support.PrintInfo("   Then: git init (in the correct folder)")
	}
}

func CheckGitInDesktop() {
	desktop := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
	if _, err := os.Stat(filepath.Join(desktop, ".git")); err == nil {
		support.PrintWarning("⚠️ .git folder found on Desktop!")
		support.PrintInfo("   Move your project to Documents/vs code/")
	}
}

func CheckGitInDownloads() {
	downloads := filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
	if _, err := os.Stat(filepath.Join(downloads, ".git")); err == nil {
		support.PrintWarning("⚠️ .git folder found in Downloads!")
		support.PrintInfo("   Move your project to Documents/vs code/")
	}
}
func CheckGitInDocuments() {
	docs := filepath.Join(os.Getenv("USERPROFILE"), "Documents")
	gitPath := filepath.Join(docs, ".git")

	if _, err := os.Stat(gitPath); err == nil {
		// This is actually okay if it's a project
		// But we can still check if it's the root
		support.PrintInfo("ℹ️ .git found in Documents (may be a project)")
	}
}
func CheckNodeModulesSize() {
	nodeModules := filepath.Join(".", "node_modules")
	if _, err := os.Stat(nodeModules); err == nil {
		size := GetDirSize(nodeModules)
		if size > 500*1024*1024 { // 500MB
			support.PrintWarning(fmt.Sprintf("⚠️ node_modules is %.1f GB", float64(size)/(1024*1024*1024)))
			support.PrintInfo("   Run: npm prune or yarn autoclean")
		} else if size > 100*1024*1024 {
			support.PrintInfo(fmt.Sprintf("ℹ️ node_modules size: %.1f MB", float64(size)/(1024*1024)))
		}
	}
}
func CheckDiskSpace() {
	cmd := exec.Command("wmic", "logicaldisk", "where", "DeviceID='C:'", "get", "FreeSpace")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 1 {
		freeStr := strings.TrimSpace(lines[1])
		freeBytes, err := strconv.ParseUint(freeStr, 10, 64)
		if err == nil {
			freeGB := float64(freeBytes) / (1024 * 1024 * 1024)
			fmt.Printf("✅ %.1f GB free\n", freeGB)
		}
	}
}
func CheckCommonEnvVars() {
	vars := []string{"GOPATH", "GOROOT", "PATH", "HOME"}
	for _, v := range vars {
		if val := os.Getenv(v); val != "" {
			if v == "PATH" && len(val) > 500 {
				support.PrintInfo(fmt.Sprintf("ℹ️ %s is set (length: %d)", v, len(val)))
			} else if val != "" {
				support.PrintSuccess(fmt.Sprintf("✅ %s is set", v))
			}
		}
	}
}
func CheckLargeFilesInGit() {
	repoRoot, err := isGitRepo()
	if err != nil || !repoRoot {
		return
	}

	cmd := exec.Command("git", "ls-files", "-s")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	largeFiles := 0
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 4 {
			size, _ := strconv.Atoi(parts[3])
			if size > 10*1024*1024 { // 10MB
				largeFiles++
			}
		}
	}

	if largeFiles > 0 {
		support.PrintWarning(fmt.Sprintf("⚠️ %d large files tracked in git (>10MB)", largeFiles))
		support.PrintInfo("   Consider using git-lfs or .gitignore")
	}
}
func CheckVSCodeExtensions() {
	homeDir, _ := os.UserHomeDir()
	extDir := filepath.Join(homeDir, ".vscode", "extensions")

	if _, err := os.Stat(extDir); err == nil {
		entries, _ := os.ReadDir(extDir)
		extCount := len(entries)

		if extCount > 50 {
			support.PrintWarning(fmt.Sprintf("⚠️ %d VS Code extensions installed", extCount))
			support.PrintInfo("   Too many extensions may slow down VS Code")
		} else {
			support.PrintSuccess(fmt.Sprintf("✅ %d extensions installed", extCount))
		}
	}
}
func CheckOmarchyVersion() {
	cmd := exec.Command("omarchy", "--version")
	output, err := cmd.Output()
	if err != nil {
		support.PrintWarning("⚠️ Could not check Omarchy version")
		return
	}

	version := strings.TrimSpace(string(output))
	support.PrintSuccess(fmt.Sprintf("✅ Omarchy %s", version))

	// Check if update available (simplified)
	support.PrintInfo("   Run 'omarchy update' to check for updates")
}
func GetDirSize(path string) int64 {
	var size int64
	filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, _ := d.Info()
			size += info.Size()
		}
		return nil
	})
	return size
}
func HandleFixGitInHome() {
	homeDir, _ := os.UserHomeDir()
	gitPath := filepath.Join(homeDir, ".git")

	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		fmt.Println("✅ No .git folder found in home directory.")
		return
	}

	fmt.Println("⚠️ Found .git folder in home directory!")
	fmt.Println("   This will be deleted.")
	fmt.Print("   Are you sure? (y/N): ")

	var resp string
	fmt.Scanln(&resp)
	if resp != "y" && resp != "Y" {
		fmt.Println("Aborted.")
		return
	}

	err := os.RemoveAll(gitPath)
	if err != nil {
		fmt.Printf("❌ Failed to remove: %v\n", err)
		return
	}

	fmt.Println("✅ Removed .git from home directory.")
	fmt.Println("   You can now git init in the correct project folder.")
}
func CheckUntrackedFiles() {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	untracked := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(untracked) > 0 && untracked[0] != "" {
		support.PrintWarning(fmt.Sprintf("⚠️ %d untracked files in git", len(untracked)))
		support.PrintInfo("   Run 'git status' to see untracked files")
	}
}
