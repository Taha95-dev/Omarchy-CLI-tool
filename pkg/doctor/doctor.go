package doctor

import (
	"context"
	"fmt"
	"omarchy/pkg/gitsupport"
	"omarchy/pkg/support"
	"os"
	"sync"
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
	fmt.Printf("🔍 Omarchy Environment Check (Thread-Safe Concurrency)\n\n")

	var wg sync.WaitGroup
	var mu sync.Mutex // 👈 This locks down os.Stdout so strings don't smash together

	// Group 1: CLI External Tools Validation
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Do your tool checks, but wrap them in a lock so they print sequentially
		mu.Lock()
		support.CheckTool("node", "--version")
		support.CheckTool("go", "version")
		support.CheckTool("git", "--version")
		support.CheckTool("npm", "--version")
		support.CheckTool("docker", "--version")
		mu.Unlock()
	}()

	// Group 2: Configuration & Git Core Status
	wg.Add(1)
	go func() {
		defer wg.Done()

		configPath := support.GetConfigPath()

		mu.Lock()
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("✅ Omarchy config found at: %s\n", configPath)
		} else {
			fmt.Printf("⚠️ Omarchy config not found (run: omarchy config init)\n")
		}
		gitsupport.CheckGitConfig()
		mu.Unlock()
	}()

	// Group 3: File System Location Safety
	wg.Add(1)
	go func() {
		defer wg.Done()

		mu.Lock()
		gitsupport.CheckGitInHome()
		gitsupport.CheckGitInDesktop()
		gitsupport.CheckGitInDownloads()
		gitsupport.CheckGitInDocuments()
		mu.Unlock()
	}()

	// Group 4: Heavy I/O Tasks (Disk space and node_modules size)
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Let the hard drive work in the background without holding the lock!
		// This is where you save time.
		gitsupport.CheckNodeModulesSize()
		gitsupport.CheckDiskSpace()
	}()

	// Group 5: Repository Health and Environment Data
	wg.Add(1)
	go func() {
		defer wg.Done()

		mu.Lock()
		gitsupport.CheckCommonEnvVars()
		gitsupport.CheckLargeFilesInGit()
		gitsupport.CheckUntrackedFiles()
		mu.Unlock()
	}()

	// Group 6: IDE Sync & Framework Versions
	wg.Add(1)
	go func() {
		defer wg.Done()

		mu.Lock()
		gitsupport.CheckVSCodeExtensions()
		gitsupport.CheckOmarchyVersion()
		mu.Unlock()
	}()

	wg.Wait()

	fmt.Println("\n✨ All diagnostic health checks complete without text racing!")
}
