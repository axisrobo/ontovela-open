// Package testutil provides a tiny fake ONTOVELA core for adapter tests.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

// FakeCore accepts assertion appends and counts them.
type FakeCore struct {
	mu    sync.Mutex
	server *httptest.Server
	count int
}

// NewFakeCore starts an httptest server accepting POST /v1/assertions.
func NewFakeCore() *FakeCore {
	core := &FakeCore{}
	core.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input ontovela.StateAssertionInput
		_ = json.NewDecoder(r.Body).Decode(&input)
		core.mu.Lock()
		core.count++
		core.mu.Unlock()
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1", TenantID: input.SubjectID})
	}))
	return core
}

// URL returns the fake core base URL.
func (c *FakeCore) URL() string { return c.server.URL }

// Count returns the number of accepted appends.
func (c *FakeCore) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// Close shuts the server down.
func (c *FakeCore) Close() { c.server.Close() }
