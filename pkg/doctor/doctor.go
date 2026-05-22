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

	fmt.Printf("Operating System: %s", support.GetOS())
	fmt.Printf("Home Directory: %s", support.GetHomeDir())
	support.PrintInfo("Omarchy Environment Check")

	// Check tools (these could also be made cancellable)
	support.CheckTool("node", "--version")
	support.CheckTool("go", "version")
	support.CheckTool("git", "--version")
	support.CheckTool("npm", "--version")
	support.CheckTool("docker", "--version")

	// Check config
	configPath := support.ConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Omarchy config found at: %s", support.ConfigPath())
	} else {
		support.PrintWarning("Omarchy config not found (run: omarchy config init)")
	}

	// Check Git config
	support.PrintInfo("\n📋 Git Configuration:")
	gitsupport.CheckGitConfig()
}
