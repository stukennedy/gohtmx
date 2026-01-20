# GoHTMX

A hypermedia-driven mobile application framework that uses Go as a runtime kernel with HTMX in a WebView. Build native iOS and Android apps using Go, HTML, and HTMX - no JavaScript frameworks required.

## Key Features

- **Go-Powered Mobile Apps**: Write your backend logic in Go, compile to native mobile frameworks
- **HTMX for Interactivity**: Use HTMX's hypermedia approach instead of complex JavaScript
- **Virtual HTTP**: No network sockets - requests are intercepted and handled directly by Go
- **Type-Safe Templates**: Use [templ](https://templ.guide) for compile-time checked HTML templates
- **Hot Reload Development**: Edit Go/templ code and see changes instantly in the iOS Simulator
- **Single Codebase**: Share business logic between iOS, Android, and web

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Mobile App                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                 WebView (HTMX)                         │  │
│  │  • HTML rendered by Go templates                       │  │
│  │  • HTMX handles interactions via gohtmx:// scheme     │  │
│  └──────────────────────┬────────────────────────────────┘  │
│                         │                                    │
│  ┌──────────────────────▼────────────────────────────────┐  │
│  │           Native Bridge (Swift / Kotlin)               │  │
│  │  • Intercepts gohtmx:// requests                       │  │
│  │  • Routes to Go via gomobile                           │  │
│  └──────────────────────┬────────────────────────────────┘  │
│                         │                                    │
│  ┌──────────────────────▼────────────────────────────────┐  │
│  │              Go Runtime (gomobile bind)                │  │
│  │  • HTTP router (chi-based)                             │  │
│  │  • Template rendering (templ)                          │  │
│  │  • Business logic                                      │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile): `go install golang.org/x/mobile/cmd/gomobile@latest && gomobile init`
- [templ](https://templ.guide): `go install github.com/a-h/templ/cmd/templ@latest`
- [air](https://github.com/air-verse/air): `go install github.com/air-verse/air@latest`
- [entr](https://github.com/eradman/entr): `brew install entr` (macOS)
- For iOS: Xcode with iOS Simulator
- For Android: Android Studio with SDK and emulator

### Install GoHTMX CLI

```bash
go install github.com/stukennedy/gohtmx/cmd/gohtmx@latest
```

Or build from source:

```bash
git clone https://github.com/stukennedy/gohtmx.git
cd gohtmx/cmd/gohtmx
go install .
```

### Create a New Project

```bash
gohtmx new myapp
cd myapp
go mod tidy
bun install  # or: npm install
```

### Development with Hot Reload

For the fastest development experience, use dev mode which connects the iOS Simulator to a local server with hot reload:

```bash
gohtmx run ios --dev
```

This will:
1. Start the development server with hot reload
2. Build and launch the iOS app in the Simulator
3. Automatically reload the app when you edit Go/templ files

### Production Build

Build a standalone app with embedded Go runtime:

```bash
gohtmx run ios
```

## Project Structure

```
myapp/
├── main.go              # Entry point, dev server setup
├── go.mod               # Go module definition
├── .air.toml            # Air hot reload configuration
├── package.json         # Node dependencies (Tailwind CSS)
├── tailwind.config.js   # Tailwind configuration
│
├── app/
│   └── app.go           # Router setup and app configuration
│
├── handlers/
│   └── handlers.go      # HTTP handlers (business logic)
│
├── templates/
│   ├── layout.templ     # Base HTML layout
│   ├── home.templ       # Home page template
│   └── components.templ # Reusable components
│
├── static/
│   ├── css/
│   │   ├── input.css    # Tailwind source
│   │   └── output.css   # Generated CSS
│   └── js/
│       └── htmx.min.js  # HTMX library
│
├── mobile/
│   └── mobile.go        # Mobile bridge setup
│
├── ios/
│   └── Example/         # Xcode project
│       └── Example/
│           ├── GoHTMXWebViewController.swift
│           ├── GoHTMXSchemeHandler.swift
│           └── GoHTMXBridge.swift
│
└── build/
    └── ios/
        └── Gohtmx.xcframework  # Built Go framework
```

## Writing Handlers

Handlers process requests and return HTML fragments. They use a simple context-based API:

```go
// handlers/handlers.go
package handlers

import (
    "myapp/templates"
    "github.com/stukennedy/gohtmx/pkg/render"
    "github.com/stukennedy/gohtmx/pkg/router"
)

func Mount(r *router.Router, renderer *render.TemplRenderer) {
    // Simple page handler
    r.GET("/about", func(ctx *router.Context) (string, error) {
        return renderer.Render(templates.AboutPage())
    })

    // Handler with URL parameters
    r.GET("/users/{id}", func(ctx *router.Context) (string, error) {
        userID := ctx.Param("id")
        user, err := fetchUser(userID)
        if err != nil {
            return renderer.Render(templates.ErrorMessage("User not found"))
        }
        return renderer.Render(templates.UserProfile(user))
    })

    // Form submission handler
    r.POST("/todos", func(ctx *router.Context) (string, error) {
        title := ctx.FormValue("title")
        todo := createTodo(title)

        // Return just the new todo item fragment
        return renderer.Render(templates.TodoItem(todo))
    })

    // Handler that triggers HTMX events
    r.DELETE("/todos/{id}", func(ctx *router.Context) (string, error) {
        id := ctx.Param("id")
        deleteTodo(id)

        // Trigger a refresh of the todo list
        ctx.Trigger("todosUpdated")
        return "", nil
    })
}
```

## Writing Templates

Templates use [templ](https://templ.guide), a type-safe HTML templating language for Go:

```go
// templates/layout.templ
package templates

templ Layout(title string) {
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8"/>
        <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
        <title>{ title }</title>
        <link rel="stylesheet" href="/static/css/output.css"/>
        <script src="/static/js/htmx.min.js"></script>
    </head>
    <body class="bg-gray-100">
        { children... }
    </body>
    </html>
}

// templates/home.templ
package templates

templ HomePage() {
    @Layout("Home") {
        <main class="container mx-auto p-4">
            <h1 class="text-2xl font-bold">Welcome to GoHTMX</h1>

            // HTMX-powered navigation
            <nav hx-boost="true">
                <a href="/about" class="text-blue-500">About</a>
                <a href="/todos" class="text-blue-500">Todos</a>
            </nav>
        </main>
    }
}

// templates/todos.templ
package templates

type Todo struct {
    ID    string
    Title string
    Done  bool
}

templ TodoList(todos []Todo) {
    <div id="todo-list">
        for _, todo := range todos {
            @TodoItem(todo)
        }
    </div>
}

templ TodoItem(todo Todo) {
    <div id={ "todo-" + todo.ID } class="flex items-center gap-2 p-2">
        <input
            type="checkbox"
            checked?={ todo.Done }
            hx-post={ "/todos/" + todo.ID + "/toggle" }
            hx-target={ "#todo-" + todo.ID }
            hx-swap="outerHTML"
        />
        <span class={ templ.KV("line-through", todo.Done) }>
            { todo.Title }
        </span>
        <button
            hx-delete={ "/todos/" + todo.ID }
            hx-target={ "#todo-" + todo.ID }
            hx-swap="delete"
            class="text-red-500"
        >
            Delete
        </button>
    </div>
}

templ AddTodoForm() {
    <form
        hx-post="/todos"
        hx-target="#todo-list"
        hx-swap="beforeend"
        hx-on::after-request="this.reset()"
        class="flex gap-2"
    >
        <input
            type="text"
            name="title"
            placeholder="New todo..."
            class="border rounded px-2 py-1"
            required
        />
        <button type="submit" class="bg-blue-500 text-white px-4 py-1 rounded">
            Add
        </button>
    </form>
}
```

## HTMX Patterns

### Navigation with hx-boost

```html
<nav hx-boost="true">
    <a href="/page1">Page 1</a>
    <a href="/page2">Page 2</a>
</nav>
```

### Loading States

```html
<button
    hx-post="/slow-action"
    hx-indicator="#spinner"
>
    Submit
</button>
<span id="spinner" class="htmx-indicator">Loading...</span>
```

### Infinite Scroll

```html
<div
    hx-get="/items?page=2"
    hx-trigger="revealed"
    hx-swap="afterend"
>
    Loading more...
</div>
```

### Form Validation

```go
templ FormWithValidation() {
    <form hx-post="/submit" hx-target="#result">
        <input
            type="email"
            name="email"
            hx-post="/validate/email"
            hx-trigger="blur"
            hx-target="next .error"
        />
        <span class="error text-red-500"></span>
        <button type="submit">Submit</button>
    </form>
    <div id="result"></div>
}
```

## Router API

The router is based on [chi](https://github.com/go-chi/chi) with a simplified API:

```go
r := router.New()

// HTTP methods
r.GET("/path", handler)
r.POST("/path", handler)
r.PUT("/path", handler)
r.DELETE("/path", handler)
r.PATCH("/path", handler)

// URL parameters
r.GET("/users/{id}", func(ctx *router.Context) (string, error) {
    id := ctx.Param("id")  // Get URL parameter
    return "", nil
})

// Query parameters
r.GET("/search", func(ctx *router.Context) (string, error) {
    q := ctx.Query("q")  // Get query parameter
    return "", nil
})

// Form values
r.POST("/submit", func(ctx *router.Context) (string, error) {
    name := ctx.FormValue("name")
    return "", nil
})

// Route groups
r.Group("/api", func(r *router.Router) {
    r.GET("/users", listUsers)
    r.POST("/users", createUser)
})

// Middleware
r.Use(loggingMiddleware)

// Static files
r.Static("/static", "static")
```

## Context Helpers

```go
func handler(ctx *router.Context) (string, error) {
    // Check if request is from HTMX
    if ctx.IsHTMX() {
        // Return fragment only
    }

    // Get request headers
    auth := ctx.Header("Authorization")

    // Set response headers
    ctx.SetHeader("X-Custom", "value")

    // Trigger HTMX events
    ctx.Trigger("itemCreated")
    ctx.TriggerWithData("itemCreated", map[string]any{"id": 123})

    // Redirect (works with HTMX)
    ctx.Redirect("/new-location")

    // Return HTML
    return "<div>Hello</div>", nil
}
```

## CLI Commands

```bash
# Create new project
gohtmx new myapp
gohtmx new .              # Initialize in current directory

# Development
gohtmx dev                # Start dev server with hot reload (web)
gohtmx run ios --dev      # Hot reload with iOS Simulator
gohtmx run android --dev  # Hot reload with Android Emulator (coming soon)

# Production builds
gohtmx build ios          # Build iOS framework
gohtmx build android      # Build Android AAR
gohtmx build all          # Build all platforms

gohtmx run ios            # Build and run on iOS Simulator
gohtmx run android        # Build and run on Android Emulator

# Utilities
gohtmx templ              # Generate templ files
gohtmx install-tools      # Install required dev tools
gohtmx version            # Print version
gohtmx help [command]     # Show help
```

## Environment Variables

- `GOHTMX_PATH`: Path to GoHTMX source (for local development)
- `GOHTMX_DEV_SERVER`: Dev server URL (set automatically in dev mode)

## Styling with Tailwind CSS

Projects include Tailwind CSS by default:

```bash
# Build CSS once
bun run css

# Watch for changes
bun run css:watch
```

Configure in `tailwind.config.js`:

```js
module.exports = {
  content: [
    './templates/**/*.templ',
    './static/**/*.js',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

## Mobile-Specific Considerations

### Safe Areas

Handle iOS notch and home indicator:

```css
body {
    padding-top: env(safe-area-inset-top);
    padding-bottom: env(safe-area-inset-bottom);
    padding-left: env(safe-area-inset-left);
    padding-right: env(safe-area-inset-right);
}
```

### Touch Interactions

HTMX works great with touch. For better mobile UX:

```html
<!-- Larger touch targets -->
<button class="p-4 min-h-[44px]" hx-post="/action">
    Tap Me
</button>

<!-- Disable double-tap zoom on interactive elements -->
<style>
button, a, input { touch-action: manipulation; }
</style>
```

### Viewport Meta

Already included in the layout template:

```html
<meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover"/>
```

## Debugging

### iOS Simulator

1. Open Safari
2. Develop menu → Simulator → Your App
3. Use Web Inspector to debug HTML/JS/Network

### Console Logging

```go
// In Go handlers
fmt.Println("Debug:", someValue)  // Shows in terminal

// In templates (renders to page)
<script>console.log("Debug:", { someValue })</script>
```

### Network Requests

In dev mode, all requests go through HTTP and can be inspected in Safari Web Inspector.

In production mode, requests use the `gohtmx://` scheme and are handled internally.

## Troubleshooting

### "Module not found" errors

```bash
go mod tidy
```

### Hot reload not working

1. Check if air is running (look for rebuild messages)
2. Verify `.air.toml` doesn't exclude your files
3. Make sure `_templ.go` is NOT in `exclude_regex`

### iOS build fails

```bash
# Ensure gomobile is initialized
gomobile init

# Check Xcode command line tools
xcode-select --install
```

### Port 8080 already in use

```bash
# Find and kill the process
lsof -i :8080
kill <PID>
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [HTMX](https://htmx.org) - The hypermedia approach that makes this possible
- [templ](https://templ.guide) - Type-safe HTML templating for Go
- [chi](https://github.com/go-chi/chi) - Lightweight Go router
- [gomobile](https://pkg.go.dev/golang.org/x/mobile) - Go on mobile platforms
- [air](https://github.com/air-verse/air) - Live reload for Go
