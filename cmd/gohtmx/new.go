package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed templates/*
var templateFS embed.FS

// getGoVersion returns the current Go version (e.g., "1.24.12")
func getGoVersion() string {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "1.23"
	}
	// Parse "go version go1.24.12 darwin/arm64"
	re := regexp.MustCompile(`go(\d+\.\d+(?:\.\d+)?)`)
	match := re.FindStringSubmatch(string(out))
	if len(match) > 1 {
		return match[1]
	}
	return "1.23"
}

func newProject(name string) error {
	// Determine project directory
	var projectDir string
	if name == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting current directory: %w", err)
		}
		projectDir = cwd
		name = filepath.Base(cwd)
	} else {
		projectDir = name
	}

	// Check if directory exists and is not empty
	if name != "." {
		if _, err := os.Stat(projectDir); err == nil {
			entries, _ := os.ReadDir(projectDir)
			if len(entries) > 0 {
				return fmt.Errorf("directory %s already exists and is not empty", projectDir)
			}
		}
	}

	fmt.Printf("Creating new gohtmx project: %s\n", name)

	// Create project structure
	dirs := []string{
		"handlers",
		"templates",
		"static/css",
		"static/js",
	}

	for _, dir := range dirs {
		path := filepath.Join(projectDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", path, err)
		}
	}

	// Copy template files
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root templates directory
		if path == "templates" {
			return nil
		}

		// Get relative path from templates/
		relPath := strings.TrimPrefix(path, "templates/")
		destPath := filepath.Join(projectDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Read template file
		content, err := templateFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		// Replace placeholders
		contentStr := string(content)
		contentStr = strings.ReplaceAll(contentStr, "{{PROJECT_NAME}}", name)
		contentStr = strings.ReplaceAll(contentStr, "{{MODULE_PATH}}", "github.com/"+name)
		contentStr = strings.ReplaceAll(contentStr, "{{GO_VERSION}}", getGoVersion())

		// Handle .tmpl extension (remove it)
		if strings.HasSuffix(destPath, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
			relPath = strings.TrimSuffix(relPath, ".tmpl")
		}

		// Write file
		if err := os.WriteFile(destPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", destPath, err)
		}

		fmt.Printf("  created: %s\n", relPath)
		return nil
	})

	if err != nil {
		return fmt.Errorf("copying templates: %w", err)
	}

	// Make scripts executable
	scripts := []string{"dev.sh"}
	for _, script := range scripts {
		path := filepath.Join(projectDir, script)
		if err := os.Chmod(path, 0755); err != nil {
			// Ignore if file doesn't exist
			continue
		}
	}

	fmt.Println()
	fmt.Println("Project created successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", name)
	fmt.Println("  go mod tidy")
	fmt.Println("  bun install        # or: npm install")
	fmt.Println("  ./dev.sh           # start development server")
	fmt.Println()

	return nil
}
