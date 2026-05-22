package support

import (
	"fmt"
	"os"
	"os/exec"
)

func PrintError(msg string) {
	fmt.Println("❌", msg)
}

func PrintWarning(msg string) {
	fmt.Println("⚠️", msg)
}

func PrintSuccess(msg string) {
	fmt.Println("✅", msg)
}
func PrintSuccessf(format string, args ...interface{}) {
	fmt.Printf("✅ "+format+"\n", args...)
}
func PrintInfo(msg string) {
	fmt.Println("ℹ️", msg)
}
func RunCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("⚠️ Warning:", err)
	}
}
