package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

func TestSecurityPageView(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
		Authz: config.AuthzConfig{Enabled: true},
	}

	sp := NewSecurityPage(cfg)
	view := sp.View()
	if !strings.Contains(view, "authn") {
		t.Error("view should show 'authn'")
	}
	if !strings.Contains(view, "authz") {
		t.Error("view should show 'authz'")
	}
}

func TestSecurityPageEscAtRoot(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
	}

	sp := NewSecurityPage(cfg)
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc at root should return a cmd")
	}
	// Execute the cmd to verify it returns goBackMsg.
	msg := cmd()
	if _, ok := msg.(goBackMsg); !ok {
		t.Errorf("expected goBackMsg, got %T", msg)
	}
}

func TestSecurityPageDrillAndEsc(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Security = config.SecurityConfig{
		Authn: config.AuthnConfig{Enabled: true},
		Authz: config.AuthzConfig{Enabled: false},
	}

	sp := NewSecurityPage(cfg)

	// Drill into authn (first item).
	sp.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Now esc should go back within editor (not send goBackMsg).
	_, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEscape})
	// After drilling in and pressing esc, should go back to editor root, no goBackMsg.
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(goBackMsg); ok {
			t.Error("esc inside editor should not send goBackMsg")
		}
	}
}

func TestSecurityPageNoConfig(t *testing.T) {
	cfg := config.NewDefault()
	// Security is zero-value, so toYAMLNode produces a node.
	// But let's test the page still renders.
	sp := NewSecurityPage(cfg)
	view := sp.View()
	if view == "" {
		t.Error("view should not be empty")
	}
}

func TestSecurityPageNilEditor(t *testing.T) {
	cfg := config.NewDefault()
	sp := &SecurityPage{cfg: cfg, editor: nil}
	view := sp.View()
	if !strings.Contains(view, "No security configuration") {
		t.Error("nil editor should show 'No security configuration'")
	}
	if !strings.Contains(view, "esc: back") {
		t.Error("nil editor should show back hint")
	}
}

func TestSecurityPageNilEditorUpdate(t *testing.T) {
	cfg := config.NewDefault()
	sp := &SecurityPage{cfg: cfg, editor: nil}
	result, cmd := sp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if result != sp {
		t.Error("should return same page")
	}
	if cmd != nil {
		t.Error("should return nil cmd")
	}
}

func makeGateWithSettings(yamlStr string) config.ServiceConfig {
	var node yaml.Node
	yaml.Unmarshal([]byte(yamlStr), &node)
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return config.ServiceConfig{Enabled: true, Host: "localhost", Port: 8084, Settings: *node.Content[0]}
	}
	return config.ServiceConfig{Enabled: true, Host: "localhost", Port: 8084}
}

func TestNewSecurityPageWithOAuth2(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Services[model.Gate] = makeGateWithSettings(`
spring:
  security:
    oauth2:
      client:
        clientId: my-client
`)
	sp := NewSecurityPage(cfg)
	view := sp.View()
	if !strings.Contains(view, "oauth2") {
		t.Error("security page should contain 'oauth2' from gate spring.security settings")
	}
}

func TestNewSecurityPageWithSSL(t *testing.T) {
	cfg := config.NewDefault()
	cfg.Services[model.Gate] = makeGateWithSettings(`
server:
  ssl:
    enabled: true
    keyStore: /path/to/keystore
`)
	sp := NewSecurityPage(cfg)
	view := sp.View()
	if !strings.Contains(view, "ssl") {
		t.Error("security page should contain 'ssl' from gate server.ssl settings")
	}
}

func TestExtractNestedMap(t *testing.T) {
	var node yaml.Node
	yaml.Unmarshal([]byte("spring:\n  security:\n    oauth2:\n      clientId: abc"), &node)
	result := extractNestedMap(&node, "spring", "security")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// The result should be a mapping node containing "oauth2".
	if result.Kind != yaml.MappingNode {
		t.Errorf("expected MappingNode, got kind=%d", result.Kind)
	}
	found := false
	for i := 0; i+1 < len(result.Content); i += 2 {
		if result.Content[i].Value == "oauth2" {
			found = true
		}
	}
	if !found {
		t.Error("expected to find 'oauth2' key in result")
	}
}

func TestExtractNestedMapMissing(t *testing.T) {
	var node yaml.Node
	yaml.Unmarshal([]byte("spring:\n  boot:\n    version: 3"), &node)
	result := extractNestedMap(&node, "spring", "security")
	if result != nil {
		t.Error("expected nil for missing key path")
	}
}

func TestExtractNestedMapNonMapping(t *testing.T) {
	// Create a scalar node, not a mapping.
	node := &yaml.Node{Kind: yaml.ScalarNode, Value: "hello"}
	result := extractNestedMap(node, "spring")
	if result != nil {
		t.Error("expected nil for non-mapping node")
	}
}
