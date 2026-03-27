package model

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestServiceNameString(t *testing.T) {
	tests := []struct {
		name     ServiceName
		expected string
	}{
		{Clouddriver, "clouddriver"},
		{Orca, "orca"},
		{Gate, "gate"},
		{Front50, "front50"},
		{Echo, "echo"},
		{Igor, "igor"},
		{Fiat, "fiat"},
		{Rosco, "rosco"},
		{Kayenta, "kayenta"},
		{Deck, "deck"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.name.String(); got != tt.expected {
				t.Errorf("ServiceName.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestServiceNameFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected ServiceName
		wantErr  bool
	}{
		{"clouddriver", Clouddriver, false},
		{"gate", Gate, false},
		{"GATE", Gate, false},
		{"unknown", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ServiceNameFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServiceNameFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ServiceNameFromString(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAllServiceNames(t *testing.T) {
	names := AllServiceNames()
	if len(names) != 10 {
		t.Errorf("AllServiceNames() returned %d names, want 10", len(names))
	}
}

func TestServiceNamePackageName(t *testing.T) {
	tests := []struct {
		name     ServiceName
		expected string
	}{
		{Clouddriver, "spinnaker-clouddriver"},
		{Deck, "spinnaker-deck"},
		{Front50, "spinnaker-front50"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.name.PackageName(); got != tt.expected {
				t.Errorf("PackageName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestServiceNameSystemdUnit(t *testing.T) {
	if got := Gate.SystemdUnit(); got != "gate.service" {
		t.Errorf("SystemdUnit() = %q, want %q", got, "gate.service")
	}
}

func TestServiceNameConfigFile(t *testing.T) {
	if got := Gate.ConfigFile(); got != "gate.yml" {
		t.Errorf("ConfigFile() = %q, want %q", got, "gate.yml")
	}
}

func TestDeploymentOrder(t *testing.T) {
	order := DeploymentOrder()
	front50Idx, fiatIdx := -1, -1
	for i, tier := range order {
		for _, s := range tier {
			if s == Front50 {
				front50Idx = i
			}
			if s == Fiat {
				fiatIdx = i
			}
		}
	}
	if front50Idx >= fiatIdx {
		t.Error("front50 must be deployed before fiat")
	}
	gateIdx := -1
	for i, tier := range order {
		for _, s := range tier {
			if s == Gate {
				gateIdx = i
			}
		}
	}
	deckIdx := -1
	for i, tier := range order {
		for _, s := range tier {
			if s == Deck {
				deckIdx = i
			}
		}
	}
	if gateIdx >= deckIdx {
		t.Error("gate must be deployed before deck")
	}
}

func TestServiceNameStringUnknown(t *testing.T) {
	unknown := ServiceName(999)
	got := unknown.String()
	if got != "ServiceName(999)" {
		t.Errorf("String() = %q, want %q", got, "ServiceName(999)")
	}
}

func TestServiceNameMarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		svc      ServiceName
		want     string
		wantErr  bool
	}{
		{"known", Gate, "gate", false},
		{"clouddriver", Clouddriver, "clouddriver", false},
		{"unknown", ServiceName(999), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.svc.MarshalYAML()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("MarshalYAML() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestServiceNameUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ServiceName
		wantErr bool
	}{
		{"gate", "gate", Gate, false},
		{"orca", "orca", Orca, false},
		{"unknown", "nonexistent", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var svc ServiceName
			err := yaml.Unmarshal([]byte(tt.input), &svc)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && svc != tt.want {
				t.Errorf("UnmarshalYAML() = %v, want %v", svc, tt.want)
			}
		})
	}
}

func TestServiceNameUnmarshalYAMLInvalidType(t *testing.T) {
	var svc ServiceName
	// Try to unmarshal an integer, which should fail since it expects a string.
	err := yaml.Unmarshal([]byte("123"), &svc)
	// This should either error on ServiceNameFromString or work depending on yaml parsing
	// The yaml library parses "123" as string "123" so ServiceNameFromString will error.
	if err == nil {
		t.Error("expected error unmarshaling '123' as ServiceName")
	}
}

func TestServiceNameYAMLRoundTrip(t *testing.T) {
	type wrapper struct {
		Service ServiceName `yaml:"service"`
	}
	original := wrapper{Service: Front50}
	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var decoded wrapper
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.Service != original.Service {
		t.Errorf("round-trip: got %v, want %v", decoded.Service, original.Service)
	}
}

func TestAllServiceNamesContents(t *testing.T) {
	names := AllServiceNames()
	expected := map[ServiceName]bool{
		Clouddriver: true, Orca: true, Gate: true, Front50: true, Echo: true,
		Igor: true, Fiat: true, Rosco: true, Kayenta: true, Deck: true,
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected service name: %v", n)
		}
		delete(expected, n)
	}
	if len(expected) > 0 {
		t.Errorf("missing service names: %v", expected)
	}
}
