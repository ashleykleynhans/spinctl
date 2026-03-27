package deploy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const testBOMYAML = `version: "1.35.0"
timestamp: "2025-01-15 00:00:00"
services:
  clouddriver:
    version: "5.82.1"
  orca:
    version: "8.47.0"
`

func TestFetchBOMFromNetwork(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testBOMYAML))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	fetcher := NewBOMFetcher(server.URL+"/%s.yml", cacheDir)

	bom, err := fetcher.Fetch("1.35.0")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if bom.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", bom.Version, "1.35.0")
	}

	// Verify it was cached.
	cached := filepath.Join(cacheDir, "1.35.0.yml")
	if _, err := os.Stat(cached); os.IsNotExist(err) {
		t.Error("expected BOM to be cached on disk")
	}
}

func TestFetchBOMFromCache(t *testing.T) {
	serverCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	cachedPath := filepath.Join(cacheDir, "1.35.0.yml")
	if err := os.WriteFile(cachedPath, []byte(testBOMYAML), 0644); err != nil {
		t.Fatal(err)
	}

	fetcher := NewBOMFetcher(server.URL+"/%s.yml", cacheDir)

	bom, err := fetcher.Fetch("1.35.0")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if bom.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", bom.Version, "1.35.0")
	}

	if serverCalled {
		t.Error("server should not have been called when cache is valid")
	}
}

func TestFetchBOMCorruptCacheRefetches(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testBOMYAML))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	cachedPath := filepath.Join(cacheDir, "1.35.0.yml")
	// Write an empty file to simulate corruption.
	if err := os.WriteFile(cachedPath, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	fetcher := NewBOMFetcher(server.URL+"/%s.yml", cacheDir)

	bom, err := fetcher.Fetch("1.35.0")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	if bom.Version != "1.35.0" {
		t.Errorf("version = %q, want %q", bom.Version, "1.35.0")
	}
}

func TestLoadFromCacheCorruptYAML(t *testing.T) {
	cacheDir := t.TempDir()
	cachedPath := filepath.Join(cacheDir, "bad.yml")
	// Write valid YAML but missing version field (parseBOMBytes requires version).
	if err := os.WriteFile(cachedPath, []byte("services:\n  orca:\n    version: 1.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	fetcher := NewBOMFetcher("http://unused/%s.yml", cacheDir)
	_, err := fetcher.loadFromCache(cachedPath)
	if err == nil {
		t.Error("expected error for BOM without version field")
	}
	// Corrupt cache should be removed.
	if _, statErr := os.Stat(cachedPath); !os.IsNotExist(statErr) {
		t.Error("corrupt cached file should be removed")
	}
}

func TestFetchBOMNetworkFailNoCacheErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	fetcher := NewBOMFetcher(server.URL+"/%s.yml", cacheDir)

	_, err := fetcher.Fetch("99.99.99")
	if err == nil {
		t.Fatal("expected error when network fails and no cache, got nil")
	}
}
