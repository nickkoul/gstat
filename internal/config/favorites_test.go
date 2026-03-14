package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestFavoritesStore(baseDir string) *FavoritesStore {
	return &FavoritesStore{
		userConfigDir: func() (string, error) {
			return baseDir, nil
		},
	}
}

func TestFavoritesStoreLoadMissingFile(t *testing.T) {
	store := newTestFavoritesStore(t.TempDir())

	favorites, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(favorites) != 0 {
		t.Fatalf("Load() favorites len = %d, want 0", len(favorites))
	}
}

func TestFavoritesStoreSaveAndLoadRoundTrip(t *testing.T) {
	baseDir := t.TempDir()
	store := newTestFavoritesStore(baseDir)

	input := map[string]bool{
		"rory":    true,
		"scottie": true,
		"ignore":  false,
		"":        true,
	}
	if err := store.Save(input); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	path := filepath.Join(baseDir, "gstat", favoritesConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	plain := string(data)
	if !strings.Contains(plain, "scottie") || !strings.Contains(plain, "rory") {
		t.Fatalf("saved file missing player IDs: %s", plain)
	}
	if strings.Contains(plain, "ignore") {
		t.Fatalf("saved file should skip false favorites: %s", plain)
	}

	favorites, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if len(favorites) != 2 || !favorites["scottie"] || !favorites["rory"] {
		t.Fatalf("Load() favorites = %#v, want scottie+rory", favorites)
	}
}

func TestFavoritesStoreLoadCorruptFile(t *testing.T) {
	baseDir := t.TempDir()
	store := newTestFavoritesStore(baseDir)
	path := filepath.Join(baseDir, "gstat", favoritesConfigFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	favorites, err := store.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if len(favorites) != 0 {
		t.Fatalf("Load() favorites len = %d, want 0 on corrupt file", len(favorites))
	}
	if !strings.Contains(err.Error(), "parse favorites") {
		t.Fatalf("Load() error = %v, want parse favorites context", err)
	}
}

func TestFavoritesStoreLoadReadError(t *testing.T) {
	baseDir := t.TempDir()
	store := newTestFavoritesStore(baseDir)
	path := filepath.Join(baseDir, "gstat", favoritesConfigFile)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}

	favorites, err := store.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want read error")
	}
	if len(favorites) != 0 {
		t.Fatalf("Load() favorites len = %d, want 0 on read error", len(favorites))
	}
	if !strings.Contains(err.Error(), "read favorites") {
		t.Fatalf("Load() error = %v, want read favorites context", err)
	}
}

func TestFavoritesStoreSaveCreateDirectoryError(t *testing.T) {
	baseDir := t.TempDir()
	blocked := filepath.Join(baseDir, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	store := newTestFavoritesStore(blocked)

	err := store.Save(map[string]bool{"scottie": true})
	if err == nil {
		t.Fatal("Save() error = nil, want directory creation error")
	}
	if !strings.Contains(err.Error(), "create config directory") {
		t.Fatalf("Save() error = %v, want config directory context", err)
	}
}

func TestFavoritesStoreLoadUserConfigDirError(t *testing.T) {
	store := &FavoritesStore{
		userConfigDir: func() (string, error) {
			return "", errors.New("boom")
		},
	}

	favorites, err := store.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want config dir error")
	}
	if len(favorites) != 0 {
		t.Fatalf("Load() favorites len = %d, want 0", len(favorites))
	}
	if !strings.Contains(err.Error(), "resolve user config directory") {
		t.Fatalf("Load() error = %v, want config dir context", err)
	}
}
