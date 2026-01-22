// Shared application logic for the todo app
package main

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/stukennedy/irgo/examples/todo/templates"
	"github.com/stukennedy/irgo/pkg/render"
	"github.com/stukennedy/irgo/pkg/router"
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

func setupRouter() *router.Router {
	r := router.New()

	// Serve static files (CSS, JS)
	r.Static("/static", http.Dir("static"))

	// Home page - renders full page with all todos
	r.GET("/", func(ctx *router.Context) (string, error) {
		todos := store.All()
		return renderer.Render(templates.HomePage(todos))
	})

	// Add new todo (Datastar SSE)
	r.DSPost("/todos", func(ctx *router.Context) error {
		var signals struct {
			Title string `json:"title"`
		}
		if err := ctx.ReadSignals(&signals); err != nil {
			signals.Title = ctx.FormValue("title")
		}

		if signals.Title == "" {
			return ctx.SSE().PatchTempl(templates.ErrorMessage("Title is required"))
		}

		todo := store.Add(signals.Title)
		sse := ctx.SSE()

		// Prepend new todo to list
		sse.PatchTempl(templates.TodoItem(todo))

		// Clear the input
		sse.PatchSignals(map[string]any{"title": ""})

		// Remove empty state if it exists
		sse.Remove("#empty-state")

		return nil
	})

	// Toggle todo completion (Datastar SSE)
	r.DSPost("/todos/{id}/toggle", func(ctx *router.Context) error {
		id := parseID(ctx.Param("id"))
		todo := store.Toggle(id)
		if todo == nil {
			ctx.NotFound("Todo not found")
			return nil
		}
		return ctx.SSE().PatchTempl(templates.TodoItem(todo))
	})

	// Delete todo (Datastar SSE)
	r.DSDelete("/todos/{id}", func(ctx *router.Context) error {
		id := parseID(ctx.Param("id"))
		store.Delete(id)

		// Remove the element from DOM
		return ctx.SSE().Remove(fmt.Sprintf("#todo-%d", id))
	})

	return r
}

func addSampleData() {
	store.Add("Learn irgo framework")
	store.Add("Build a mobile app with Datastar")
	store.Add("Deploy to iOS and Android")
}

func parseID(s string) int64 {
	var id int64
	fmt.Sscanf(s, "%d", &id)
	return id
}
