package model

import "testing"

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
