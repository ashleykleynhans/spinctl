package tui

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestToYAMLNodeMap(t *testing.T) {
	input := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	node, err := toYAMLNode(input)
	if err != nil {
		t.Fatalf("toYAMLNode() error = %v", err)
	}
	if node == nil {
		t.Fatal("toYAMLNode() returned nil node")
	}
	if node.Kind != yaml.MappingNode {
		t.Errorf("node.Kind = %v, want MappingNode", node.Kind)
	}
	// Map with 2 keys should have 4 content nodes (key, value pairs).
	if len(node.Content) != 4 {
		t.Errorf("node.Content len = %d, want 4", len(node.Content))
	}
}

func TestToYAMLNodeNil(t *testing.T) {
	node, err := toYAMLNode(nil)
	if err != nil {
		t.Fatalf("toYAMLNode(nil) error = %v", err)
	}
	if node == nil {
		t.Fatal("toYAMLNode(nil) returned nil node")
	}
}

func TestToYAMLNodeScalar(t *testing.T) {
	node, err := toYAMLNode("hello")
	if err != nil {
		t.Fatalf("toYAMLNode() error = %v", err)
	}
	if node == nil {
		t.Fatal("toYAMLNode() returned nil node")
	}
	if node.Kind != yaml.ScalarNode {
		t.Errorf("node.Kind = %v, want ScalarNode", node.Kind)
	}
}

func TestToYAMLNodeStruct(t *testing.T) {
	type testStruct struct {
		Name    string `yaml:"name"`
		Enabled bool   `yaml:"enabled"`
		Port    int    `yaml:"port"`
	}

	input := testStruct{
		Name:    "test",
		Enabled: true,
		Port:    8080,
	}

	node, err := toYAMLNode(input)
	if err != nil {
		t.Fatalf("toYAMLNode() error = %v", err)
	}
	if node == nil {
		t.Fatal("toYAMLNode() returned nil node")
	}
	if node.Kind != yaml.MappingNode {
		t.Errorf("node.Kind = %v, want MappingNode", node.Kind)
	}
	// Struct with 3 fields should have 6 content nodes.
	if len(node.Content) != 6 {
		t.Errorf("node.Content len = %d, want 6", len(node.Content))
	}
}

