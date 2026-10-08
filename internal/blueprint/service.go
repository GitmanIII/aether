package blueprint

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aether/internal/platform"
)

// Service manages blueprint persistence.
type Service struct {
	dir string
}

// NewService creates a new blueprint service.
func NewService() *Service {
	dir := platform.BlueprintDir()
	if err := platform.EnsureDir(dir); err != nil {
		log.Printf("[blueprint] ensure dir %s: %v", dir, err)
	}
	return &Service{dir: dir}
}

// LoadAll reads all blueprints from the blueprints directory.
func (s *Service) LoadAll() ([]Blueprint, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read blueprints dir: %w", err)
	}

	var blueprints []Blueprint
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		bp, err := s.loadFromFile(path, entry.Name())
		if err != nil {
			continue
		}
		blueprints = append(blueprints, bp)
	}
	return blueprints, nil
}

// FindByName finds a blueprint by name (case-insensitive).
func (s *Service) FindByName(name string) (*Blueprint, error) {
	blueprints, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	lower := strings.ToLower(name)

	// Exact name match first
	for i := range blueprints {
		if strings.ToLower(blueprints[i].Name) == lower {
			return &blueprints[i], nil
		}
	}

	// Partial filename match
	for i := range blueprints {
		fn := strings.ToLower(strings.TrimSuffix(blueprints[i].Filename, ".json"))
		if strings.Contains(fn, lower) {
			return &blueprints[i], nil
		}
	}

	return nil, nil
}

// normalizeName returns the canonical blueprint identity: trimmed and
// lowercased. Blueprint names are case-insensitive, so "Cyber3" and "cyber3"
// are the same blueprint. This matches Omarchy, which lowercases theme names
// on install.
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// safeFilename turns a normalized blueprint name into a filename stem. Path
// separators and spaces become hyphens; the result is already lowercase.
func safeFilename(name string) string {
	safe := strings.ReplaceAll(name, "/", "-")
	return strings.ReplaceAll(safe, " ", "-")
}

// Save persists a blueprint to disk. The name is normalized before it is
// written, so saving "Cyber3" and "cyber3" updates one blueprint instead of
// creating two case variants.
func (s *Service) Save(name string, bp Blueprint) error {
	name = normalizeName(name)
	bp.Name = name
	bp.Timestamp = time.Now().UnixMilli()
	if err := validateBlueprint(&bp); err != nil {
		return fmt.Errorf("validate blueprint: %w", err)
	}

	path := filepath.Join(s.dir, safeFilename(name)+".json")
	return platform.WriteJSON(path, bp)
}

// Delete removes a blueprint by its case-insensitive name. Because identity is
// case-insensitive, it also removes legacy case variants (for example
// Cyber3.json and cyber3.json) that predate name normalization. Exact-name
// matching is used; fuzzy lookup is never used for deletion.
func (s *Service) Delete(name string) error {
	target := normalizeName(name)
	if target == "" {
		return fmt.Errorf("blueprint name must not be empty")
	}
	blueprints, err := s.LoadAll()
	if err != nil {
		return err
	}
	var matches []Blueprint
	for i := range blueprints {
		if normalizeName(blueprints[i].Name) == target {
			matches = append(matches, blueprints[i])
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("blueprint %q not found", name)
	}
	for _, match := range matches {
		if err := os.Remove(match.Path); err != nil {
			return fmt.Errorf("delete blueprint %q: %w", name, err)
		}
	}
	return nil
}

// Validate checks a blueprint's structure and color values.
func (s *Service) Validate(bp *Blueprint) bool {
	return validateBlueprint(bp) == nil
}

func (s *Service) loadFromFile(path, filename string) (Blueprint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Blueprint{}, err
	}

	var bp Blueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return Blueprint{}, err
	}

	bp.Path = path
	bp.Filename = filename
	if bp.Name == "" {
		bp.Name = strings.TrimSuffix(filename, ".json")
	}
	if err := validateBlueprint(&bp); err != nil {
		return Blueprint{}, err
	}

	return bp, nil
}
