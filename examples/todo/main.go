// Example: Todo app demonstrating gohtmx framework usage with templ
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/stukennedy/gohtmx/examples/todo/templates"
	"github.com/stukennedy/gohtmx/mobile"
	"github.com/stukennedy/gohtmx/pkg/render"
	"github.com/stukennedy/gohtmx/pkg/router"
)

// TodoStore is a simple in-memory store
type TodoStore struct {
	todos   map[int64]*templates.Todo
	counter int64
	mu      sync.RWMutex
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos: make(map[int64]*templates.Todo),
	}
}

func (s *TodoStore) Add(title string) *templates.Todo {
	id := atomic.AddInt64(&s.counter, 1)
	todo := &templates.Todo{ID: id, Title: title}

	s.mu.Lock()
	s.todos[id] = todo
	s.mu.Unlock()

	return todo
}

func (s *TodoStore) Get(id int64) *templates.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.todos[id]
}

func (s *TodoStore) All() []*templates.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*templates.Todo, 0, len(s.todos))
	for _, t := range s.todos {
		result = append(result, t)
	}
	return result
}

func (s *TodoStore) Toggle(id int64) *templates.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.todos[id]; ok {
		t.Completed = !t.Completed
		return t
	}
	return nil
}

func (s *TodoStore) Delete(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.todos, id)
}

// Global store and renderer
var (
	store    = NewTodoStore()
	renderer = render.NewTemplRenderer()
)

func main() {
	// Check if running as desktop dev server or mobile initialization
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runDevServer()
		return
	}

	// Mobile mode: initialize bridge
	initMobile()
}

// initMobile sets up the framework for mobile use
func initMobile() {
	mobile.Initialize()

	r := setupRouter()
	mobile.SetHandler(r.Handler())

	// Add sample data
	addSampleData()

	fmt.Println("Todo app initialized for mobile")
}

// runDevServer starts an HTTP server for desktop testing
func runDevServer() {
	r := setupRouter()

	// Add sample data
	addSampleData()

	port := ":8080"
	fmt.Printf("Starting dev server at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, r.Handler()))
}

func setupRouter() *router.Router {
	r := router.New()

	// Home page - renders full page with all todos
	r.GET("/", func(ctx *router.Context) (string, error) {
		todos := store.All()
		return renderer.Render(templates.HomePage(todos))
	})

	// Get all todos (fragment)
	r.GET("/todos", func(ctx *router.Context) (string, error) {
		todos := store.All()
		return renderer.Render(templates.TodoList(todos))
	})

	// Add new todo
	r.POST("/todos", func(ctx *router.Context) (string, error) {
		title := ctx.FormValue("title")
		if title == "" {
			return renderer.Render(templates.ErrorMessage("Title is required"))
		}

		todo := store.Add(title)
		return renderer.Render(templates.TodoItem(todo))
	})

	// Toggle todo completion
	r.POST("/todos/{id}/toggle", func(ctx *router.Context) (string, error) {
		id := parseID(ctx.Param("id"))
		todo := store.Toggle(id)
		if todo == nil {
			ctx.NotFound("Todo not found")
			return "", nil
		}
		return renderer.Render(templates.TodoItem(todo))
	})

	// Delete todo
	r.DELETE("/todos/{id}", func(ctx *router.Context) (string, error) {
		id := parseID(ctx.Param("id"))
		store.Delete(id)
		// Return empty to remove the element
		return "", nil
	})

	return r
}

func addSampleData() {
	store.Add("Learn gohtmx framework")
	store.Add("Build a mobile app with HTMX")
	store.Add("Deploy to iOS and Android")
}

func parseID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
