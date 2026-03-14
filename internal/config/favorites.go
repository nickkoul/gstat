package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const favoritesConfigFile = "favorites.json"

type favoritesFile struct {
	PlayerIDs []string `json:"player_ids"`
}

// FavoritesStore persists favorited player IDs under the user's config directory.
type FavoritesStore struct {
	userConfigDir func() (string, error)
}

// NewFavoritesStore creates a store backed by os.UserConfigDir.
func NewFavoritesStore() *FavoritesStore {
	return &FavoritesStore{userConfigDir: os.UserConfigDir}
}

// Load returns the stored favorites map. Missing files are treated as empty state.
func (s *FavoritesStore) Load() (map[string]bool, error) {
	path, err := s.path()
	if err != nil {
		return map[string]bool{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]bool{}, nil
		}
		return map[string]bool{}, fmt.Errorf("read favorites: %w", err)
	}

	if len(data) == 0 {
		return map[string]bool{}, nil
	}

	var file favoritesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return map[string]bool{}, fmt.Errorf("parse favorites: %w", err)
	}

	favorites := make(map[string]bool, len(file.PlayerIDs))
	for _, playerID := range file.PlayerIDs {
		if playerID == "" {
			continue
		}
		favorites[playerID] = true
	}

	return favorites, nil
}

// Save writes favorites atomically, creating the config directory when needed.
func (s *FavoritesStore) Save(favorites map[string]bool) error {
	path, err := s.path()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	playerIDs := make([]string, 0, len(favorites))
	for playerID, favorite := range favorites {
		if favorite && playerID != "" {
			playerIDs = append(playerIDs, playerID)
		}
	}
	sort.Strings(playerIDs)

	data, err := json.MarshalIndent(favoritesFile{PlayerIDs: playerIDs}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal favorites: %w", err)
	}
	data = append(data, '\n')

	tempFile, err := os.CreateTemp(dir, "favorites-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp favorites file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("write temp favorites file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp favorites file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace favorites file: %w", err)
	}

	return nil
}

func (s *FavoritesStore) path() (string, error) {
	baseDir, err := s.userConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}
	return filepath.Join(baseDir, "gstat", favoritesConfigFile), nil
}
