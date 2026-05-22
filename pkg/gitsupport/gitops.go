package gitsupport

import (
	"context"
	"fmt"
	"omarchy/pkg/support"
	"os/exec"
	"strings"
	"time"
)

func GitSync(ctx context.Context, autoMsg bool, customMsg string) {
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
