package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"

	"github.com/spinnaker/spinctl/internal/config"
	"github.com/spinnaker/spinctl/internal/model"
)

// SecurityPage displays authentication and authorization settings
// including gate's OAuth2/SAML/LDAP configuration from service settings.
type SecurityPage struct {
	cfg    *config.SpinctlConfig
	editor *EditorPage
}

// NewSecurityPage creates a security settings page that combines
// the top-level security config with gate's spring security settings.
func NewSecurityPage(cfg *config.SpinctlConfig) *SecurityPage {
	// Build a combined security view.
	combined := make(map[string]any)

	// Add top-level authn/authz.
	combined["authn"] = map[string]any{"enabled": cfg.Security.Authn.Enabled}
	combined["authz"] = map[string]any{"enabled": cfg.Security.Authz.Enabled}

	// Extract gate's spring.security settings if present.
	if gateSvc, ok := cfg.Services[model.Gate]; ok && gateSvc.Settings.Kind != 0 {
		springSettings := extractNestedMap(&gateSvc.Settings, "spring", "security")
		if springSettings != nil {
			// Marshal the spring.security node back to map for inclusion.
			var springMap map[string]any
			data, _ := yaml.Marshal(springSettings)
			yaml.Unmarshal(data, &springMap)
			if springMap != nil {
				for k, v := range springMap {
					combined[k] = v
				}
			}
		}

		// Also check for gate's SSL settings.
		sslSettings := extractNestedMap(&gateSvc.Settings, "server", "ssl")
		if sslSettings != nil {
			var sslMap map[string]any
			data, _ := yaml.Marshal(sslSettings)
			yaml.Unmarshal(data, &sslMap)
			if sslMap != nil {
				combined["ssl"] = sslMap
			}
		}
	}

	node, _ := toYAMLNode(combined)
	var editor *EditorPage
	if node != nil {
		editor = NewEditorPage(node, "Security")
	}
	return &SecurityPage{cfg: cfg, editor: editor}
}

// extractNestedMap walks a yaml.Node tree following the given keys
// and returns the final node if found.
func extractNestedMap(node *yaml.Node, keys ...string) *yaml.Node {
	current := node
	// Unwrap document node.
	if current.Kind == yaml.DocumentNode && len(current.Content) > 0 {
		current = current.Content[0]
	}
	for _, key := range keys {
		if current.Kind != yaml.MappingNode {
			return nil
		}
		found := false
		for i := 0; i+1 < len(current.Content); i += 2 {
			if current.Content[i].Value == key {
				current = current.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return current
}

func (s *SecurityPage) Update(msg tea.Msg) (page, tea.Cmd) {
	if s.editor != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "esc" && len(s.editor.nodeStack) == 0 {
				return s, func() tea.Msg { return goBackMsg{} }
			}
		}
		var cmd tea.Cmd
		_, cmd = s.editor.Update(msg)
		return s, cmd
	}
	return s, nil
}

func (s *SecurityPage) View() string {
	if s.editor != nil {
		return s.editor.View()
	}
	var b strings.Builder
	b.WriteString("\n  Security\n\n")
	b.WriteString("  No security configuration.\n")
	b.WriteString("\n  esc: back\n")
	return b.String()
}
