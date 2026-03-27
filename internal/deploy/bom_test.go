package deploy

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spinnaker/spinctl/internal/model"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func TestParseBOM(t *testing.T) {
	bom, err := parseBOMFile(testdataPath("bom_1.35.0.yaml"))
	if err != nil {
		t.Fatalf("parseBOMFile: %v", err)
	}

	if bom.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", bom.Version, "1.35.0")
	}

	v, err := bom.ServiceVersion(model.Clouddriver)
	if err != nil {
		t.Fatalf("ServiceVersion(clouddriver): %v", err)
	}
	if v != "5.82.1" {
		t.Errorf("clouddriver version = %q, want %q", v, "5.82.1")
	}

	v, err = bom.ServiceVersion(model.Deck)
	if err != nil {
		t.Fatalf("ServiceVersion(deck): %v", err)
	}
	if v != "3.16.0" {
		t.Errorf("deck version = %q, want %q", v, "3.16.0")
	}
}

func TestBOMServiceVersionNotFound(t *testing.T) {
	bom := &BOM{
		Version:  "1.0.0",
		Services: map[string]BOMService{},
	}

	_, err := bom.ServiceVersion(model.Orca)
	if err == nil {
		t.Fatal("expected error for missing service, got nil")
	}
}

func TestResolveVersionsWithOverrides(t *testing.T) {
	bom, err := parseBOMFile(testdataPath("bom_1.35.0.yaml"))
	if err != nil {
		t.Fatalf("parseBOMFile: %v", err)
	}

	overrides := map[model.ServiceName]string{
		model.Orca: "8.99.0-custom",
	}

	services := []model.ServiceName{model.Orca, model.Gate}
	versions, err := ResolveVersions(bom, overrides, services)
	if err != nil {
		t.Fatalf("ResolveVersions: %v", err)
	}

	if versions[model.Orca] != "8.99.0-custom" {
		t.Errorf("orca version = %q, want %q", versions[model.Orca], "8.99.0-custom")
	}
	if versions[model.Gate] != "6.62.0" {
		t.Errorf("gate version = %q, want %q", versions[model.Gate], "6.62.0")
	}
}
