// CLI tool for creating and managing gohtmx projects
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gohtmx new <project-name>")
			os.Exit(1)
		}
		err = newProject(os.Args[2])

	case "dev":
		err = runDev()

	case "serve":
		err = runServe()

	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: gohtmx build <ios|android|all>")
			os.Exit(1)
		}
		err = runBuild(os.Args[2])

	case "templ":
		err = runTempl()

	case "test":
		err = runTest()

	case "version", "-v", "--version":
		fmt.Printf("gohtmx %s\n", version)

	case "help", "-h", "--help":
		if len(os.Args) > 2 {
			printCommandHelp(os.Args[2])
		} else {
			printUsage()
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`gohtmx - Hypermedia framework for mobile apps

Usage:
  gohtmx <command> [arguments]

Commands:
  new <name>       Create a new gohtmx project
  dev              Run development server with hot reload
  serve            Run server without file watching
  build <target>   Build for mobile (ios, android, or all)
  templ            Generate templ files
  test             Run tests
  version          Print version information
  help [command]   Show help for a command

Examples:
  gohtmx new myapp       Create a new project
  gohtmx dev             Start dev server with hot reload
  gohtmx build ios       Build iOS framework
  gohtmx build android   Build Android AAR
  gohtmx build all       Build all mobile platforms`)
}

func printCommandHelp(cmd string) {
	switch cmd {
	case "new":
		fmt.Println(`gohtmx new - Create a new gohtmx project

Usage:
  gohtmx new <project-name>
  gohtmx new .              Initialize in current directory

Creates a new project with:
  - main.go           App entry point
  - handlers/         Route handlers
  - templates/        Templ templates
  - static/           CSS and JS assets
  - dev.sh            Development script
  - Makefile          Build targets`)

	case "dev":
		fmt.Println(`gohtmx dev - Run development server with hot reload

Usage:
  gohtmx dev

Starts:
  - Air for Go hot reloading
  - Templ file watcher
  - Tailwind CSS watcher (if configured)

Server runs at http://localhost:8080`)

	case "build":
		fmt.Println(`gohtmx build - Build for mobile platforms

Usage:
  gohtmx build ios       Build iOS framework (.xcframework)
  gohtmx build android   Build Android library (.aar)
  gohtmx build all       Build all platforms

Requirements:
  - iOS: Xcode and gomobile
  - Android: Android SDK and gomobile

Output:
  - iOS: build/ios/Gohtmx.xcframework
  - Android: build/android/gohtmx.aar`)

	case "templ":
		fmt.Println(`gohtmx templ - Generate templ files

Usage:
  gohtmx templ

Runs 'templ generate' to compile .templ files to Go code.`)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}
