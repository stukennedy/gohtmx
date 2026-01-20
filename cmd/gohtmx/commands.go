package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// runDev starts the development server with hot reload
func runDev() error {
	// Check for required tools
	if err := checkTool("air", "go install github.com/air-verse/air@latest"); err != nil {
		return err
	}
	if err := checkTool("templ", "go install github.com/a-h/templ/cmd/templ@latest"); err != nil {
		return err
	}
	if err := checkTool("entr", "brew install entr"); err != nil {
		return err
	}

	// Check if dev.sh exists (user project) or we're in framework
	if _, err := os.Stat("dev.sh"); err == nil {
		// User project - run dev.sh
		return runCommand("./dev.sh")
	}

	// Framework development - run air directly
	fmt.Println("Starting development server...")

	// Generate templ files first
	if err := runTempl(); err != nil {
		fmt.Printf("Warning: templ generate failed: %v\n", err)
	}

	return runCommand("air")
}

// runServe starts the server without file watching
func runServe() error {
	// Check if main.go exists
	if _, err := os.Stat("main.go"); err == nil {
		// User project
		return runCommand("go", "run", ".", "serve")
	}

	// Framework - run example
	if _, err := os.Stat("examples/todo/main.go"); err == nil {
		return runCommand("go", "run", "./examples/todo", "serve")
	}

	return fmt.Errorf("no main.go found - are you in a gohtmx project?")
}

// runBuild builds for mobile platforms
func runBuild(target string) error {
	// Check for gomobile
	if err := checkTool("gomobile", "go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init"); err != nil {
		return err
	}

	// Determine module path
	modulePath, err := getModulePath()
	if err != nil {
		return fmt.Errorf("could not determine module path: %w", err)
	}

	// Create build directory
	if err := os.MkdirAll("build", 0755); err != nil {
		return fmt.Errorf("creating build directory: %w", err)
	}

	switch target {
	case "ios":
		return buildIOS(modulePath)
	case "android":
		return buildAndroid(modulePath)
	case "all":
		if err := buildIOS(modulePath); err != nil {
			return err
		}
		return buildAndroid(modulePath)
	default:
		return fmt.Errorf("unknown build target: %s (use ios, android, or all)", target)
	}
}

func buildIOS(modulePath string) error {
	fmt.Println("Building iOS framework...")

	outPath := "build/ios/Gohtmx.xcframework"
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	// Remove existing framework
	os.RemoveAll(outPath)

	mobilePackage := modulePath + "/mobile"
	if err := runCommand("gomobile", "bind", "-target", "ios", "-o", outPath, mobilePackage); err != nil {
		return fmt.Errorf("gomobile bind failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("iOS framework built: %s\n", outPath)
	fmt.Println()
	fmt.Println("To use in Xcode:")
	fmt.Printf("  1. Drag %s into your Xcode project\n", outPath)
	fmt.Println("  2. Add to 'Frameworks, Libraries, and Embedded Content'")
	fmt.Println("  3. Import Gohtmx in your Swift code")

	return nil
}

func buildAndroid(modulePath string) error {
	fmt.Println("Building Android AAR...")

	outPath := "build/android/gohtmx.aar"
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}

	// Remove existing AAR
	os.Remove(outPath)

	mobilePackage := modulePath + "/mobile"
	if err := runCommand("gomobile", "bind", "-target", "android", "-o", outPath, mobilePackage); err != nil {
		return fmt.Errorf("gomobile bind failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("Android AAR built: %s\n", outPath)
	fmt.Println()
	fmt.Println("To use in Android Studio:")
	fmt.Println("  1. Copy to app/libs/gohtmx.aar")
	fmt.Println("  2. Add to build.gradle: implementation files('libs/gohtmx.aar')")

	return nil
}

// runTempl generates templ files
func runTempl() error {
	if err := checkTool("templ", "go install github.com/a-h/templ/cmd/templ@latest"); err != nil {
		return err
	}

	fmt.Println("Generating templ files...")
	return runCommand("templ", "generate")
}

// runTest runs the test suite
func runTest() error {
	fmt.Println("Running tests...")
	return runCommand("go", "test", "-v", "./...")
}

// Helper functions

func checkTool(name, installCmd string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s not found. Install with: %s", name, installCmd)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func getModulePath() (string, error) {
	// Try to read from go.mod
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	// Parse module line
	lines := string(data)
	for _, line := range splitLines(lines) {
		if len(line) > 7 && line[:7] == "module " {
			return line[7:], nil
		}
	}

	return "", fmt.Errorf("module directive not found in go.mod")
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
