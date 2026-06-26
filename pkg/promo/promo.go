package promo

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	promoFile = ".omarchy_promo"
)

type Message struct {
	Tip     string
	Link    string
	Tagline string
}

var promos = []Message{
	{
		Tip:     "💡 Turn 'omarchy sync -a' into a 3-letter shortcut with v2",
		Link:    "https://buy.polar.sh/polar_cl_zvnuIMcqUEG0ghrgaFfFV9PivHvnI9esOA40D25wvUK",
		Tagline: "Save hours every week — $5 one-time",
	},
	{
		Tip:     "⏱️ You've typed 20+ commands today. v2 slashes that to 3-letter shortcuts.",
		Link:    "https://buy.polar.sh/polar_cl_zvnuIMcqUEG0ghrgaFfFV9PivHvnI9esOA40D25wvUK",
		Tagline: "Your fingers will thank you. $5 lifetime.",
	},
	{
		Tip:     "🔥 Create shortcuts for ANY command. Even multi-step workflows.",
		Link:    "https://buy.polar.sh/polar_cl_zvnuIMcqUEG0ghrgaFfFV9PivHvnI9esOA40D25wvUK",
		Tagline: "One payment. Lifetime updates. $5.",
	},
	{
		Tip:     "🚀 Omarchy v2 — shortcuts for git, docker, deploy, and more.",
		Link:    "https://buy.polar.sh/polar_cl_zvnuIMcqUEG0ghrgaFfFV9PivHvnI9esOA40D25wvUK",
		Tagline: "Stop typing. Start shortcutting. $5.",
	},
}

// ShouldShow checks if it's been at least 7 days since the last promo
func ShouldShow() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return true
	}

	path := filepath.Join(home, promoFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return true
	}

	lastShown, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return true
	}

	return time.Since(lastShown) >= 7*24*time.Hour
}

// UpdateTimestamp updates the last promo timestamp
func UpdateTimestamp() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	path := filepath.Join(home, promoFile)
	now := time.Now().Format(time.RFC3339)
	os.WriteFile(path, []byte(now), 0644)
}

// GetWeeklyMessage returns a random promo message
func GetWeeklyMessage() Message {
	day := time.Now().YearDay()
	index := day % len(promos)
	return promos[index]
}

// Show displays the weekly promotion if conditions are met
func Show() {
	if !ShouldShow() {
		return
	}

	// Check if user has v2 (optional: create a license check)
	if hasV2License() {
		return
	}

	msg := GetWeeklyMessage()
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  %s\n", msg.Tip)
	fmt.Printf("  👉 %s\n", msg.Link)
	fmt.Printf("  %s\n", msg.Tagline)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	UpdateTimestamp()
}

// hasV2License is a placeholder — customize this later
func hasV2License() bool {
	// Check for a license file or environment variable
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	path := filepath.Join(home, ".omarchy_v2")
	_, err = os.Stat(path)
	return err == nil
}
