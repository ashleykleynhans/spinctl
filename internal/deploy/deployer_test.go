package deploy

import (
	"context"
	"fmt"
	"sync"

	"github.com/spinnaker/spinctl/internal/model"
)

// MockExecutor records commands and optionally simulates errors.
type MockExecutor struct {
	mu       sync.Mutex
	Commands []string
	FailOn   map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		FailOn: make(map[string]error),
	}
}

func (m *MockExecutor) Run(_ context.Context, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cmd := name
	for _, a := range args {
		cmd += " " + a
	}
	m.Commands = append(m.Commands, cmd)
	if err, ok := m.FailOn[name]; ok {
		return err
	}
	return nil
}

func (m *MockExecutor) SetFail(name string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailOn[name] = err
}

func (m *MockExecutor) CommandCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Commands)
}

func (m *MockExecutor) HasCommand(substr string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.Commands {
		if containsStr(c, substr) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// errSimulated is a reusable test error.
var errSimulated = fmt.Errorf("simulated failure")

// testBOM returns a minimal BOM for testing.
func testBOM() *BOM {
	return &BOM{
		Version: "1.35.0",
		Services: map[string]BOMService{
			"clouddriver": {Version: "5.82.1"},
			"deck":        {Version: "3.16.0"},
			"echo":        {Version: "2.40.0"},
			"fiat":        {Version: "1.43.0"},
			"front50":     {Version: "2.33.0"},
			"gate":        {Version: "6.62.0"},
			"igor":        {Version: "4.18.0"},
			"kayenta":     {Version: "2.40.0"},
			"orca":        {Version: "8.47.0"},
			"rosco":       {Version: "1.20.0"},
		},
	}
}

// testServices returns a list with a single service for simple tests.
func testServices(names ...model.ServiceName) []model.ServiceName {
	return names
}
