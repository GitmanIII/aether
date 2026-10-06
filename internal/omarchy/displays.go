package omarchy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aether/internal/platform"
)

// DisplayAssignment is a per-display background chosen through a per-screen
// background service. Aether writes images; colors and gradients can appear
// when they were set by the service's own UI.
type DisplayAssignment struct {
	Type  string `json:"type"`
	Path  string `json:"path,omitempty"`
	Color string `json:"color,omitempty"`
}

// Display is one physically connected output enriched for Aether's UI.
type Display struct {
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Make           string             `json:"make"`
	Model          string             `json:"model"`
	Serial         string             `json:"serial"`
	X              int                `json:"x"`
	Y              int                `json:"y"`
	Width          int                `json:"width"`
	Height         int                `json:"height"`
	PhysicalWidth  int                `json:"physicalWidth"`
	PhysicalHeight int                `json:"physicalHeight"`
	Scale          float64            `json:"scale"`
	Transform      int                `json:"transform"`
	Portrait       bool               `json:"portrait"`
	Focused        bool               `json:"focused"`
	Key            string             `json:"key"`
	Keys           []string           `json:"keys"`
	Assignment     *DisplayAssignment `json:"assignment,omitempty"`
}

// DisplaysResult reports the connected displays and whether the running
// Omarchy shell can render a different background per display. When PerScreen
// is false, only the global background is available and per-display assignment
// methods return ErrPerScreenUnsupported.
type DisplaysResult struct {
	PerScreen bool      `json:"perScreen"`
	Displays  []Display `json:"displays"`
}

// ErrPerScreenUnsupported is returned when the active background service has
// no per-screen API. Aether integrates with any service that keeps Omarchy's
// "background" IPC target and adds setForScreen/clearForScreen (for example
// the Backdrop plugin).
var ErrPerScreenUnsupported = fmt.Errorf("no per-display background service is active")

type displayState struct {
	Version  int                          `json:"version"`
	Displays map[string]DisplayAssignment `json:"displays"`
}

// Displays lists the connected outputs and merges any per-display assignments
// from the active background service.
func Displays() (DisplaysResult, error) {
	monitors, err := platform.Monitors()
	if err != nil {
		return DisplaysResult{}, err
	}

	assignments, perScreen := backgroundAssignments()

	result := DisplaysResult{PerScreen: perScreen, Displays: []Display{}}
	for _, monitor := range monitors {
		if monitor.Disabled {
			continue
		}
		result.Displays = append(result.Displays, newDisplay(monitor, assignments))
	}
	return result, nil
}

// newDisplay builds the UI view of a monitor and attaches any assignment that
// matches one of its candidate keys.
func newDisplay(monitor platform.Monitor, assignments map[string]DisplayAssignment) Display {
	logicalW, logicalH := monitor.LogicalSize()
	physicalW, physicalH := monitor.PhysicalSize()
	keys := displayKeys(monitor)

	display := Display{
		Name:           monitor.Name,
		Description:    monitor.Description,
		Make:           monitor.Make,
		Model:          monitor.Model,
		Serial:         monitor.Serial,
		X:              monitor.X,
		Y:              monitor.Y,
		Width:          logicalW,
		Height:         logicalH,
		PhysicalWidth:  physicalW,
		PhysicalHeight: physicalH,
		Scale:          monitor.Scale,
		Transform:      monitor.Transform,
		Portrait:       monitor.Portrait(),
		Focused:        monitor.Focused,
		Keys:           keys,
	}
	if len(keys) > 0 {
		display.Key = keys[0]
	}
	for _, key := range keys {
		if assignment, ok := assignments[key]; ok {
			assignment := assignment
			display.Assignment = &assignment
			break
		}
	}
	return display
}

// displayKeys returns the assignment keys for a monitor, most stable first.
// This mirrors the background service's own lookup: a serial-backed
// make:model:serial key survives connector renumbering, with the connector
// name as a fallback.
func displayKeys(monitor platform.Monitor) []string {
	keys := []string{}
	if serial := strings.TrimSpace(monitor.Serial); serial != "" {
		parts := make([]string, 0, 3)
		for _, part := range []string{monitor.Make, monitor.Model, serial} {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				parts = append(parts, trimmed)
			}
		}
		if len(parts) > 0 {
			keys = append(keys, strings.Join(parts, ":"))
		}
	}
	if name := strings.TrimSpace(monitor.Name); name != "" {
		keys = append(keys, name)
	}
	return keys
}

// SetDisplayWallpaper assigns an image to a display through the active
// per-screen background service. Every candidate key for the display is set,
// because a service may resolve a display by its connector name rather than
// the serial-backed key.
func SetDisplayWallpaper(screenKeys []string, path string) error {
	keys := normalizeKeys(screenKeys)
	if len(keys) == 0 {
		return fmt.Errorf("a display key is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve wallpaper: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("inspect wallpaper: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("wallpaper is not a regular file")
	}
	if !PerScreenSupported() {
		return ErrPerScreenUnsupported
	}
	for _, key := range keys {
		out, err := platform.RunSync("omarchy-shell", "background", "setForScreen", key, abs)
		if err != nil {
			return fmt.Errorf("set display background: %w", err)
		}
		if strings.TrimSpace(out) == "invalid" {
			return fmt.Errorf("the background service rejected display %q", key)
		}
	}
	return nil
}

// ClearDisplayWallpaper removes a display's assignment so it falls back to the
// global Omarchy background. Every candidate key is cleared so a stale
// connector entry cannot resurface after the serial key is removed.
func ClearDisplayWallpaper(screenKeys []string) error {
	keys := normalizeKeys(screenKeys)
	if len(keys) == 0 {
		return fmt.Errorf("a display key is required")
	}
	if !PerScreenSupported() {
		return ErrPerScreenUnsupported
	}
	for _, key := range keys {
		if _, err := platform.RunSync("omarchy-shell", "background", "clearForScreen", key); err != nil {
			return fmt.Errorf("clear display background: %w", err)
		}
	}
	return nil
}

// normalizeKeys trims, drops empty and de-duplicates assignment keys while
// preserving their order.
func normalizeKeys(keys []string) []string {
	normalized := make([]string, 0, len(keys))
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	return normalized
}

// PerScreenSupported reports whether the active background service implements
// the per-screen assignment API.
func PerScreenSupported() bool {
	_, ok := backgroundAssignments()
	return ok
}

// backgroundAssignments reads the active service's assignment map. The second
// return value is false when no per-screen service is available, in which case
// the map is empty.
func backgroundAssignments() (map[string]DisplayAssignment, bool) {
	if !IsInstalled() || !platform.CommandExists("omarchy-shell") {
		return nil, false
	}
	out, err := platform.RunSync("omarchy-shell", "background", "assignments")
	if err != nil {
		return nil, false
	}
	assignments, err := parseAssignments([]byte(out))
	if err != nil {
		return nil, false
	}
	return assignments, true
}

// parseAssignments decodes the JSON returned by the background service's
// assignments() method. Unknown or malformed entries are skipped so a service
// extension cannot break the list.
func parseAssignments(data []byte) (map[string]DisplayAssignment, error) {
	var state displayState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse display assignments: %w", err)
	}
	assignments := make(map[string]DisplayAssignment, len(state.Displays))
	for key, assignment := range state.Displays {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		switch assignment.Type {
		case "image":
			if strings.TrimSpace(assignment.Path) != "" {
				assignments[key] = DisplayAssignment{Type: "image", Path: assignment.Path}
			}
		case "color":
			if strings.TrimSpace(assignment.Color) != "" {
				assignments[key] = DisplayAssignment{Type: "color", Color: assignment.Color}
			}
		case "gradient":
			assignments[key] = DisplayAssignment{Type: "gradient"}
		}
	}
	return assignments, nil
}
