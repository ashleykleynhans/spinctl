package deploy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DefaultBOMURLPattern is the default URL pattern for fetching BOMs from GCS.
const DefaultBOMURLPattern = "https://storage.googleapis.com/halconfig/bom/%s.yml"

// BOMFetcher fetches BOM files from a remote URL and caches them locally.
type BOMFetcher struct {
	urlPattern string
	cacheDir   string
}

// NewBOMFetcher creates a new BOMFetcher with the given URL pattern and cache
// directory.
func NewBOMFetcher(urlPattern, cacheDir string) *BOMFetcher {
	return &BOMFetcher{
		urlPattern: urlPattern,
		cacheDir:   cacheDir,
	}
}

// Fetch retrieves a BOM by version, checking the local cache first and falling
// back to an HTTP fetch.
func (f *BOMFetcher) Fetch(version string) (*BOM, error) {
	cachedPath := filepath.Join(f.cacheDir, version+".yml")

	bom, err := f.loadFromCache(cachedPath)
	if err == nil {
		return bom, nil
	}

	data, err := f.fetchFromNetwork(version)
	if err != nil {
		return nil, err
	}

	bom, err = parseBOMBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parsing fetched BOM: %w", err)
	}

	// Cache the fetched BOM.
	if err := os.MkdirAll(f.cacheDir, 0700); err == nil {
		_ = os.WriteFile(cachedPath, data, 0600)
	}

	return bom, nil
}

// loadFromCache attempts to load a BOM from the local cache. If the cached
// file is corrupt or empty, it is removed and an error is returned.
func (f *BOMFetcher) loadFromCache(cachedPath string) (*BOM, error) {
	data, err := os.ReadFile(cachedPath)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		_ = os.Remove(cachedPath)
		return nil, fmt.Errorf("cached BOM is empty")
	}

	bom, err := parseBOMBytes(data)
	if err != nil {
		_ = os.Remove(cachedPath)
		return nil, fmt.Errorf("cached BOM is corrupt: %w", err)
	}

	return bom, nil
}

// fetchFromNetwork downloads a BOM from the remote URL.
func (f *BOMFetcher) fetchFromNetwork(version string) ([]byte, error) {
	url := fmt.Sprintf(f.urlPattern, version)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching BOM from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching BOM: HTTP %d from %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading BOM response: %w", err)
	}

	return data, nil
}
