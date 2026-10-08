package blueprint

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadAllJSON(t *testing.T) {
	svc := newTestService(t)
	for _, name := range []string{"Night Sky", "Sunrise"} {
		if err := svc.Save(name, testBlueprint()); err != nil {
			t.Fatal(err)
		}
	}
	bps, err := svc.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}
	if len(bps) != 2 {
		t.Fatalf("loaded %d blueprints, want 2", len(bps))
	}

	// Test JSON serialization (what Wails does)
	data, err := json.Marshal(bps)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Test round-trip
	var decoded []Blueprint
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	for i := range bps {
		if bps[i].Path != filepath.Join(svc.dir, bps[i].Filename) || bps[i].Timestamp == 0 {
			t.Errorf("missing blueprint metadata: %+v", bps[i])
		}
		bps[i].Path = ""
		bps[i].Filename = ""
	}
	if !reflect.DeepEqual(decoded, bps) {
		t.Errorf("JSON round trip changed blueprints: %+v", decoded)
	}
}

func TestDeleteIsCaseInsensitive(t *testing.T) {
	for _, tt := range []struct {
		name        string
		query       string
		names       []string
		wantDeleted []string
		wantError   string
	}{
		{"empty", "", []string{"Night Sky"}, nil, "must not be empty"},
		{"whitespace", " \t\n", []string{"Night Sky"}, nil, "must not be empty"},
		{"unique substring", "Night", []string{"Night Sky"}, nil, "not found"},
		{"ambiguous substring", "Night", []string{"Night Sky", "Night Sea"}, nil, "not found"},
		{"filename is not display name", "first", []string{"Night Sky"}, nil, "not found"},
		{"unknown", "Missing", []string{"Night Sky"}, nil, "not found"},
		{"path traversal", "../first", []string{"Night Sky"}, nil, "not found"},
		{"full name", "Night Sky", []string{"Night Sky", "Night Sky Bright"}, []string{"first.json"}, ""},
		{"case insensitive", "NIGHT SKY", []string{"Night Sky", "Sunrise"}, []string{"first.json"}, ""},
		{"duplicate display names", "Night Sky", []string{"Night Sky", "Night Sky"}, []string{"first.json", "second.json"}, ""},
		{"case variants", "NIGHT SKY", []string{"Night Sky", "night sky"}, []string{"first.json", "second.json"}, ""},
		{"legacy filename fallback", "FIRST", []string{"", "Sunrise"}, []string{"first.json"}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(t)
			originals := make(map[string]string)
			for i, name := range tt.names {
				bp := testBlueprint()
				bp.Name = name
				data, err := json.Marshal(bp)
				if err != nil {
					t.Fatal(err)
				}
				filename := []string{"first.json", "second.json"}[i]
				path := filepath.Join(svc.dir, filename)
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
				originals[filename] = string(data)
			}

			err := svc.Delete(tt.query)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("Delete(%q) error = %v, want %q", tt.query, err, tt.wantError)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			deleted := make(map[string]bool, len(tt.wantDeleted))
			for _, filename := range tt.wantDeleted {
				deleted[filename] = true
			}
			for filename, original := range originals {
				data, err := os.ReadFile(filepath.Join(svc.dir, filename))
				if deleted[filename] {
					if !os.IsNotExist(err) {
						t.Errorf("%s was not deleted: %v", filename, err)
					}
				} else if err != nil || string(data) != original {
					t.Errorf("unselected blueprint %s changed: %v", filename, err)
				}
			}
		})
	}
}

// TestSaveNormalizesNameToLowerCase covers the reported bug: saving "Cyber3"
// then "cyber3" must not create two blueprints.
func TestSaveNormalizesNameToLowerCase(t *testing.T) {
	svc := newTestService(t)
	for _, name := range []string{"Cyber3", "cyber3", "CYBER3"} {
		if err := svc.Save(name, testBlueprint()); err != nil {
			t.Fatal(err)
		}
	}
	bps, err := svc.LoadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(bps) != 1 {
		t.Fatalf("loaded %d blueprints, want 1", len(bps))
	}
	if bps[0].Name != "cyber3" {
		t.Errorf("name = %q, want %q", bps[0].Name, "cyber3")
	}
	entries, err := os.ReadDir(svc.dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "cyber3.json" {
		t.Fatalf("unexpected blueprint files: %v, %v", entries, err)
	}
}

// TestDeleteRemovesLegacyCaseVariants covers blueprints created before names
// were normalized: "Cyber3.json" and "cyber3.json" must both be deletable.
func TestDeleteRemovesLegacyCaseVariants(t *testing.T) {
	svc := newTestService(t)
	for _, name := range []string{"Cyber3", "cyber3"} {
		bp := testBlueprint()
		bp.Name = name
		data, err := json.Marshal(bp)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(svc.dir, name+".json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := svc.Delete("CyBeR3"); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(svc.dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("legacy case variants not removed: %v, %v", entries, err)
	}
}

func TestFindByNameRetainsFuzzyLookup(t *testing.T) {
	svc := newTestService(t)
	if err := svc.Save("Night Sky", testBlueprint()); err != nil {
		t.Fatal(err)
	}
	bp, err := svc.FindByName("night")
	if err != nil || bp == nil || bp.Name != "night sky" {
		t.Fatalf("fuzzy lookup = %+v, %v", bp, err)
	}
}

func TestSaveFailurePreservesBlueprint(t *testing.T) {
	svc := newTestService(t)
	if err := svc.Save("Night Sky", testBlueprint()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(svc.dir, "night-sky.json")
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	unmarshalable := testBlueprint()
	unmarshalable.Adjustments = map[string]float64{"brightness": math.NaN()}
	for _, bp := range []Blueprint{{}, unmarshalable} {
		if err := svc.Save("Night Sky", bp); err == nil {
			t.Fatal("expected invalid blueprint save to fail")
		}
		data, err := os.ReadFile(path)
		if err != nil || string(data) != string(original) {
			t.Fatalf("failed save damaged blueprint: %v", err)
		}
	}
	if err := svc.Save("Night Sky", testBlueprint()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("save changed private permissions: %v, %v", info, err)
	}
	entries, err := os.ReadDir(svc.dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("unexpected blueprint directory contents: %v, %v", entries, err)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return NewService()
}

func testBlueprint() Blueprint {
	colors := make([]string, 16)
	for i := range colors {
		colors[i] = "#123456"
	}
	return Blueprint{Palette: PaletteData{Colors: colors}}
}
