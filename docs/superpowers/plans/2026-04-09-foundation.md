# Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the monorepo with a minimal Go + chi API and Vue 3 frontend, wired together with Taskfiles, multi-stage Dockerfiles, and docker-compose for local dev.

**Architecture:** Two independent services under `services/`. Each has its own `go.mod` / `package.json`, its own `Taskfile.yml`, and its own multi-stage `Dockerfile`. A root `Taskfile.yml` aggregates them. `docker-compose.yml` at root wires them for local dev, with nginx proxying `/api` to the Go backend.

**Tech Stack:** Go 1.23 + chi v5, Vue 3 + Vite 6 + Vitest 3, Taskfile v3, Docker (multi-stage, distroless Go / nginx Vue), docker compose v2

> **Note:** The Go module path uses `github.com/OWNER/dx-connect-ci-scaffold/services/api`. Replace `OWNER` with your actual GitHub username or org throughout Tasks 2–6.

---

## File Map

```
.gitignore
Taskfile.yml
docker-compose.yml
services/
  api/
    cmd/server/main.go
    internal/
      store/
        items.go
        items_test.go
      handler/
        health.go
        health_test.go
        items.go
        items_test.go
        router.go
    .golangci.yml
    Dockerfile
    Taskfile.yml
    go.mod
  web/
    src/
      components/
        ItemList.vue
        ItemList.test.js
      App.vue
      main.js
    eslint.config.js
    index.html
    nginx.conf
    Dockerfile
    Taskfile.yml
    package.json
    vite.config.js
```

---

## Task 1: Repository Foundation

**Files:**
- Create: `.gitignore`
- Create: `services/api/.gitkeep` (directory placeholder, removed in Task 2)
- Create: `services/web/.gitkeep` (directory placeholder, removed in Task 8)
- Create: `deploy/azure/.gitkeep`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p services/api services/web deploy/azure
```

- [ ] **Step 2: Write .gitignore**

```gitignore
# Go
bin/
*.exe
*.test
*.out

# Node
node_modules/
dist/
.vite/

# IDE
.vscode/
.idea/
*.swp

# OS
.DS_Store
Thumbs.db

# Superpowers brainstorm artifacts
.superpowers/

# Task runner cache
.task/
```

- [ ] **Step 3: Create deploy placeholder**

```bash
touch deploy/azure/.gitkeep
```

- [ ] **Step 4: Commit**

```bash
git add .gitignore deploy/azure/.gitkeep
git commit -m "chore: initial repository structure"
```

---

## Task 2: Go Module Setup

**Files:**
- Create: `services/api/go.mod`

- [ ] **Step 1: Initialise Go module**

Run from `services/api/`:

```bash
cd services/api
go mod init github.com/OWNER/dx-connect-ci-scaffold/services/api
```

This creates `go.mod`. Replace `OWNER` with your GitHub username or org — e.g. `github.com/acme/dx-connect-ci-scaffold/services/api`.

- [ ] **Step 2: Add chi dependency**

```bash
go get github.com/go-chi/chi/v5@latest
```

`go.mod` and `go.sum` are now present.

- [ ] **Step 3: Verify go.mod looks like**

```
module github.com/OWNER/dx-connect-ci-scaffold/services/api

go 1.23

require github.com/go-chi/chi/v5 v5.x.x
```

- [ ] **Step 4: Commit**

```bash
cd ../..
git add services/api/go.mod services/api/go.sum
git commit -m "chore(api): initialise Go module with chi"
```

---

## Task 3: ItemStore (TDD)

**Files:**
- Create: `services/api/internal/store/items.go`
- Create: `services/api/internal/store/items_test.go`

- [ ] **Step 1: Write the failing test**

`services/api/internal/store/items_test.go`:

```go
package store_test

import (
	"testing"

	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/store"
)

func TestItemStore_EmptyOnCreate(t *testing.T) {
	s := store.NewItemStore()

	items := s.List()
	if len(items) != 0 {
		t.Fatalf("expected empty store, got %d items", len(items))
	}
}

func TestItemStore_CreateAndList(t *testing.T) {
	s := store.NewItemStore()

	item := s.Create("widget")

	if item.Name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", item.Name)
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}

	items := s.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != item.ID {
		t.Errorf("expected item ID %q, got %q", item.ID, items[0].ID)
	}
}

func TestItemStore_ListReturnsCopy(t *testing.T) {
	s := store.NewItemStore()
	s.Create("widget")

	items := s.List()
	items[0].Name = "mutated"

	fresh := s.List()
	if fresh[0].Name == "mutated" {
		t.Error("List() should return a copy, not a reference to internal state")
	}
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
cd services/api
go test ./internal/store/...
```

Expected: `cannot find package` or `undefined: store`

- [ ] **Step 3: Implement ItemStore**

`services/api/internal/store/items.go`:

```go
package store

import (
	"fmt"
	"sync"
)

// Item is the domain model for a named item.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ItemStore is a thread-safe in-memory store for Items.
type ItemStore struct {
	mu    sync.Mutex
	items []Item
	seq   int
}

// NewItemStore returns an initialised empty ItemStore.
func NewItemStore() *ItemStore {
	return &ItemStore{items: []Item{}}
}

// List returns a copy of all stored items.
func (s *ItemStore) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Item, len(s.items))
	copy(result, s.items)
	return result
}

// Create adds a new item with the given name and returns it.
func (s *ItemStore) Create(name string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	item := Item{ID: fmt.Sprintf("%d", s.seq), Name: name}
	s.items = append(s.items, item)
	return item
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/store/... -v
```

Expected:
```
--- PASS: TestItemStore_EmptyOnCreate
--- PASS: TestItemStore_CreateAndList
--- PASS: TestItemStore_ListReturnsCopy
PASS
```

- [ ] **Step 5: Commit**

```bash
cd ../..
git add services/api/internal/store/
git commit -m "feat(api): add in-memory ItemStore"
```

---

## Task 4: HTTP Handlers (TDD)

**Files:**
- Create: `services/api/internal/handler/health.go`
- Create: `services/api/internal/handler/health_test.go`
- Create: `services/api/internal/handler/items.go`
- Create: `services/api/internal/handler/items_test.go`

- [ ] **Step 1: Write health handler test**

`services/api/internal/handler/health_test.go`:

```go
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/store"
)

func newTestHandler() *handler.Handler {
	return handler.New(store.NewItemStore())
}

func TestHealth_Returns200WithStatusOK(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status %q, got %q", "ok", body["status"])
	}
}
```

- [ ] **Step 2: Write items handler tests**

`services/api/internal/handler/items_test.go`:

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/store"
)

func TestListItems_ReturnsEmptyArray(t *testing.T) {
	h := handler.New(store.NewItemStore())
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	w := httptest.NewRecorder()

	h.ListItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []store.Item
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}

func TestCreateItem_Returns201WithItem(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`{"name":"widget"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var item store.Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.Name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", item.Name)
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateItem_MissingName_Returns400(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateItem_InvalidJSON_Returns400(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
```

- [ ] **Step 3: Run to confirm failure**

```bash
cd services/api
go test ./internal/handler/...
```

Expected: `undefined: handler`

- [ ] **Step 4: Implement Handler struct**

`services/api/internal/handler/health.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/store"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	items *store.ItemStore
}

// New returns a Handler wired to the given ItemStore.
func New(items *store.ItemStore) *Handler {
	return &Handler{items: items}
}

// Health responds with {"status":"ok"}.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
}
```

- [ ] **Step 5: Implement items handlers**

`services/api/internal/handler/items.go`:

```go
package handler

import (
	"encoding/json"
	"net/http"
)

type createRequest struct {
	Name string `json:"name"`
}

// ListItems responds with the full list of items as JSON.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.items.List()) //nolint:errcheck
}

// CreateItem parses a JSON body and creates a new item.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	item := h.items.Create(req.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item) //nolint:errcheck
}
```

- [ ] **Step 6: Run tests to confirm they pass**

```bash
go test ./internal/... -v
```

Expected: all tests in `store` and `handler` packages pass.

- [ ] **Step 7: Commit**

```bash
cd ../..
git add services/api/internal/handler/
git commit -m "feat(api): add health and items HTTP handlers"
```

---

## Task 5: Router and Entrypoint

**Files:**
- Create: `services/api/internal/handler/router.go`
- Create: `services/api/cmd/server/main.go`

- [ ] **Step 1: Write router**

`services/api/internal/handler/router.go`:

```go
package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter wires the Handler to a chi router and returns it.
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", h.Health)
	r.Route("/api", func(r chi.Router) {
		r.Get("/items", h.ListItems)
		r.Post("/items", h.CreateItem)
	})

	return r
}
```

- [ ] **Step 2: Write entrypoint**

`services/api/cmd/server/main.go`:

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/OWNER/dx-connect-ci-scaffold/services/api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	items := store.NewItemStore()
	h := handler.New(items)
	r := handler.NewRouter(h)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 3: Verify the binary builds**

```bash
cd services/api
go build ./...
```

Expected: no output, no errors. A binary is produced (will be cleaned up in Task 6).

- [ ] **Step 4: Smoke test the server manually (optional)**

```bash
go run ./cmd/server &
curl http://localhost:8080/health
# expected: {"status":"ok"}
curl http://localhost:8080/api/items
# expected: []
kill %1
```

- [ ] **Step 5: Run full test suite**

```bash
go test ./...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd ../..
git add services/api/internal/handler/router.go services/api/cmd/
git commit -m "feat(api): wire router and server entrypoint"
```

---

## Task 6: Go Taskfile and Linting Config

**Files:**
- Create: `services/api/Taskfile.yml`
- Create: `services/api/.golangci.yml`

- [ ] **Step 1: Write API Taskfile**

`services/api/Taskfile.yml`:

```yaml
version: '3'

tasks:
  build:
    desc: Build the API binary
    cmds:
      - go build -o bin/server ./cmd/server

  test:
    desc: Run tests
    cmds:
      - go test ./...

  test:verbose:
    desc: Run tests with verbose output
    cmds:
      - go test -v ./...

  lint:
    desc: Run golangci-lint
    cmds:
      - golangci-lint run

  audit:
    desc: Run govulncheck
    cmds:
      - govulncheck ./...

  run:
    desc: Run the server locally
    cmds:
      - go run ./cmd/server

  docker:build:
    desc: Build Docker image
    cmds:
      - docker build -t api:dev .

  docker:run:
    desc: Run Docker image on port 8080
    cmds:
      - docker run --rm -p 8080:8080 api:dev

  clean:
    desc: Remove build artifacts
    cmds:
      - rm -rf bin/
```

- [ ] **Step 2: Write golangci-lint config**

`services/api/.golangci.yml`:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - goimports
    - gosimple
    - bodyclose

linters-settings:
  goimports:
    local-prefixes: github.com/OWNER/dx-connect-ci-scaffold
```

- [ ] **Step 3: Install required tools if not present**

```bash
# golangci-lint
brew install golangci-lint
# or: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# govulncheck (used by the audit task)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

- [ ] **Step 4: Run lint to verify config is valid**

```bash
cd services/api
task lint
```

Expected: no errors (or fix any lint issues raised before continuing).

- [ ] **Step 5: Run task test to verify Taskfile works**

```bash
task test
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
cd ../..
git add services/api/Taskfile.yml services/api/.golangci.yml
git commit -m "chore(api): add Taskfile and golangci-lint config"
```

---

## Task 7: Vue Scaffold and ESLint Config

**Files:**
- Create: `services/web/package.json`
- Create: `services/web/vite.config.js`
- Create: `services/web/index.html`
- Create: `services/web/eslint.config.js`
- Create: `services/web/src/main.js`
- Create: `services/web/src/App.vue`

- [ ] **Step 1: Write package.json**

`services/web/package.json`:

```json
{
  "name": "web",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src",
    "audit": "npm audit --audit-level=moderate"
  },
  "dependencies": {
    "vue": "^3.5.0"
  },
  "devDependencies": {
    "@eslint/js": "^9.0.0",
    "@vitejs/plugin-vue": "^5.0.0",
    "@vue/test-utils": "^2.4.0",
    "eslint": "^9.0.0",
    "eslint-plugin-vue": "^9.0.0",
    "globals": "^15.0.0",
    "jsdom": "^25.0.0",
    "vite": "^6.0.0",
    "vitest": "^3.0.0"
  }
}
```

- [ ] **Step 2: Write vite.config.js**

`services/web/vite.config.js`:

```js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
  },
})
```

- [ ] **Step 3: Write index.html**

`services/web/index.html`:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>dx-connect</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.js"></script>
  </body>
</html>
```

- [ ] **Step 4: Write ESLint config (flat config, ESLint v9)**

`services/web/eslint.config.js`:

```js
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import globals from 'globals'

export default [
  js.configs.recommended,
  ...pluginVue.configs['flat/recommended'],
  {
    languageOptions: {
      globals: { ...globals.browser },
    },
  },
  {
    ignores: ['dist/', 'node_modules/'],
  },
]
```

- [ ] **Step 5: Write App.vue**

`services/web/src/App.vue`:

```vue
<template>
  <div id="app">
    <h1>Items</h1>
    <ItemList />
  </div>
</template>

<script setup>
import ItemList from './components/ItemList.vue'
</script>
```

- [ ] **Step 6: Write main.js**

`services/web/src/main.js`:

```js
import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')
```

- [ ] **Step 7: Install dependencies**

```bash
cd services/web
npm install
```

- [ ] **Step 8: Verify build works**

```bash
npm run build
```

Expected: `dist/` directory created, no errors.

- [ ] **Step 9: Run lint**

```bash
npm run lint
```

Expected: no errors (App.vue exists but ItemList.vue doesn't yet — lint may warn. That's fine for now; it will resolve in Task 8.)

- [ ] **Step 10: Commit**

```bash
cd ../..
git add services/web/package.json services/web/package-lock.json services/web/vite.config.js services/web/index.html services/web/eslint.config.js services/web/src/
git commit -m "feat(web): scaffold Vue 3 app with Vite and ESLint"
```

---

## Task 8: ItemList Component (TDD)

**Files:**
- Create: `services/web/src/components/ItemList.vue`
- Create: `services/web/src/components/ItemList.test.js`

- [ ] **Step 1: Write the failing tests**

`services/web/src/components/ItemList.test.js`:

```js
import { mount, flushPromises } from '@vue/test-utils'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ItemList from './ItemList.vue'

describe('ItemList', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
  })

  it('renders items fetched on mount', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([{ id: '1', name: 'widget' }]),
    })

    const wrapper = mount(ItemList)
    await flushPromises()

    const items = wrapper.findAll('[data-testid="item"]')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toBe('widget')
  })

  it('renders empty list when no items exist', async () => {
    global.fetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve([]),
    })

    const wrapper = mount(ItemList)
    await flushPromises()

    expect(wrapper.findAll('[data-testid="item"]')).toHaveLength(0)
  })

  it('adds an item on form submit and appends to list', async () => {
    global.fetch
      .mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
      .mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: '1', name: 'gadget' }),
      })

    const wrapper = mount(ItemList)
    await flushPromises()

    await wrapper.find('[data-testid="item-input"]').setValue('gadget')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    const items = wrapper.findAll('[data-testid="item"]')
    expect(items).toHaveLength(1)
    expect(items[0].text()).toBe('gadget')
    expect(wrapper.find('[data-testid="item-input"]').element.value).toBe('')
  })

  it('shows an error when fetch fails', async () => {
    global.fetch.mockRejectedValueOnce(new Error('network error'))

    const wrapper = mount(ItemList)
    await flushPromises()

    expect(wrapper.find('[data-testid="error"]').exists()).toBe(true)
  })
})
```

- [ ] **Step 2: Run to confirm failure**

```bash
cd services/web
npm test
```

Expected: `Cannot find module './ItemList.vue'`

- [ ] **Step 3: Implement ItemList.vue**

`services/web/src/components/ItemList.vue`:

```vue
<template>
  <div>
    <form @submit.prevent="addItem">
      <input
        v-model="newName"
        data-testid="item-input"
        placeholder="Item name"
        type="text"
      />
      <button type="submit">Add</button>
    </form>

    <ul>
      <li
        v-for="item in items"
        :key="item.id"
        data-testid="item"
      >
        {{ item.name }}
      </li>
    </ul>

    <p v-if="error" data-testid="error">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const items = ref([])
const newName = ref('')
const error = ref('')

async function fetchItems() {
  try {
    const res = await fetch('/api/items')
    items.value = await res.json()
  } catch {
    error.value = 'Failed to load items'
  }
}

async function addItem() {
  if (!newName.value.trim()) return
  try {
    const res = await fetch('/api/items', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newName.value.trim() }),
    })
    if (!res.ok) throw new Error('Failed to create item')
    const item = await res.json()
    items.value.push(item)
    newName.value = ''
  } catch (e) {
    error.value = e.message
  }
}

onMounted(fetchItems)
</script>
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
npm test
```

Expected:
```
✓ ItemList > renders items fetched on mount
✓ ItemList > renders empty list when no items exist
✓ ItemList > adds an item on form submit and appends to list
✓ ItemList > shows an error when fetch fails
```

- [ ] **Step 5: Run lint**

```bash
npm run lint
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
cd ../..
git add services/web/src/components/
git commit -m "feat(web): add ItemList component with tests"
```

---

## Task 9: Web Taskfile

**Files:**
- Create: `services/web/Taskfile.yml`

- [ ] **Step 1: Write web Taskfile**

`services/web/Taskfile.yml`:

```yaml
version: '3'

tasks:
  install:
    desc: Install dependencies
    cmds:
      - npm ci

  build:
    desc: Build the Vue app for production
    cmds:
      - npm run build

  test:
    desc: Run tests
    cmds:
      - npm run test

  test:watch:
    desc: Run tests in watch mode
    cmds:
      - npm run test:watch

  lint:
    desc: Run ESLint
    cmds:
      - npm run lint

  audit:
    desc: Run npm audit
    cmds:
      - npm audit --audit-level=moderate

  run:
    desc: Start Vite dev server (proxies /api to localhost:8080)
    cmds:
      - npm run dev

  docker:build:
    desc: Build Docker image
    cmds:
      - docker build -t web:dev .

  docker:run:
    desc: Run Docker image on port 3000
    cmds:
      - docker run --rm -p 3000:80 web:dev
```

- [ ] **Step 2: Verify tasks run**

```bash
cd services/web
task test
task lint
```

Expected: both pass.

- [ ] **Step 3: Commit**

```bash
cd ../..
git add services/web/Taskfile.yml
git commit -m "chore(web): add Taskfile"
```

---

## Task 10: Root Taskfile

**Files:**
- Create: `Taskfile.yml`

- [ ] **Step 1: Write root Taskfile**

`Taskfile.yml`:

```yaml
version: '3'

includes:
  api:
    taskfile: ./services/api/Taskfile.yml
    dir: ./services/api
  web:
    taskfile: ./services/web/Taskfile.yml
    dir: ./services/web

tasks:
  build:
    desc: Build all services
    deps: [api:build, web:build]

  test:
    desc: Test all services
    deps: [api:test, web:test]

  lint:
    desc: Lint all services
    deps: [api:lint, web:lint]

  audit:
    desc: Run security audits for all services
    deps: [api:audit, web:audit]

  up:
    desc: Start all services with docker compose
    cmds:
      - docker compose up --build

  up:detach:
    desc: Start all services in background
    cmds:
      - docker compose up --build -d

  down:
    desc: Stop all services
    cmds:
      - docker compose down
```

- [ ] **Step 2: Verify namespaced tasks work from root**

```bash
task api:test
task web:test
task test
```

Expected: all tests pass.

- [ ] **Step 3: Commit**

```bash
git add Taskfile.yml
git commit -m "chore: add root Taskfile with service includes"
```

---

## Task 11: API Dockerfile

**Files:**
- Create: `services/api/Dockerfile`

- [ ] **Step 1: Write the Dockerfile**

`services/api/Dockerfile`:

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Cache dependency download layer separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Distroless: no shell, minimal attack surface, non-root by default
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

- [ ] **Step 2: Build the image**

```bash
cd services/api
task docker:build
```

Expected: image `api:dev` built successfully.

- [ ] **Step 3: Run the image and smoke test**

```bash
task docker:run &
sleep 2
curl http://localhost:8080/health
# expected: {"status":"ok"}
curl http://localhost:8080/api/items
# expected: []
docker stop $(docker ps -q --filter ancestor=api:dev)
```

- [ ] **Step 4: Commit**

```bash
cd ../..
git add services/api/Dockerfile
git commit -m "feat(api): add multi-stage Dockerfile with distroless runtime"
```

---

## Task 12: Web Dockerfile and nginx Config

**Files:**
- Create: `services/web/Dockerfile`
- Create: `services/web/nginx.conf`
- Create: `docker-compose.yml`

- [ ] **Step 1: Write nginx.conf**

`services/web/nginx.conf`:

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # Proxy API calls to the backend service.
    # In docker-compose the backend is reachable as "api".
    # For production (ACA/AKS), replace this with the actual backend URL
    # or use a different nginx.conf via ConfigMap / volume mount.
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /health {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
    }

    # Long-lived cache for hashed asset bundles
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # SPA fallback — serve index.html for all unmatched routes
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

- [ ] **Step 2: Write web Dockerfile**

`services/web/Dockerfile`:

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app

# Cache npm install layer separately from source
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

- [ ] **Step 3: Build the web image**

```bash
cd services/web
task docker:build
```

Expected: image `web:dev` built successfully.

- [ ] **Step 4: Write docker-compose.yml**

`docker-compose.yml` (at repo root):

```yaml
services:
  api:
    build:
      context: ./services/api
    ports:
      - "8080:8080"
    environment:
      PORT: "8080"

  web:
    build:
      context: ./services/web
    ports:
      - "3000:80"
    depends_on:
      - api
```

- [ ] **Step 5: Run the full stack**

```bash
cd ../..
task up:detach
```

Wait ~10 seconds for containers to start, then:

```bash
curl http://localhost:3000/health
# expected: {"status":"ok"}  (proxied through nginx to Go API)

curl http://localhost:3000/api/items
# expected: []

curl -X POST http://localhost:3000/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"widget"}'
# expected: {"id":"1","name":"widget"}

curl http://localhost:3000/api/items
# expected: [{"id":"1","name":"widget"}]

task down
```

- [ ] **Step 6: Commit**

```bash
git add services/web/Dockerfile services/web/nginx.conf docker-compose.yml
git commit -m "feat: add web Dockerfile, nginx config, and docker-compose"
```

---

## Done

At this point the repo has:
- A working Go API (`task api:test`, `task api:lint`, `task api:build`)
- A working Vue frontend (`task web:test`, `task web:lint`, `task web:build`)
- Root-level orchestration (`task test`, `task build`, `task up`)
- Multi-stage Dockerfiles for both services
- `docker compose up` spins up a working full stack at `http://localhost:3000`

**Next:** Plan B — CI/CD Pipelines (GitHub Actions workflows, release-please, Renovate, branch protection, Azure deploy configs).
