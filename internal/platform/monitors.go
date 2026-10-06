package platform

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
)

// Monitor is one output as reported by Hyprland. Width and Height are the
// physical pixel dimensions before the transform is applied; use LogicalSize
// for the on-screen size. Transform values 1, 3, 5 and 7 rotate the output by
// 90 or 270 degrees (portrait).
type Monitor struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Make        string  `json:"make"`
	Model       string  `json:"model"`
	Serial      string  `json:"serial"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Scale       float64 `json:"scale"`
	Transform   int     `json:"transform"`
	Focused     bool    `json:"focused"`
	Disabled    bool    `json:"disabled"`
	MirrorOf    string  `json:"mirrorOf"`
}

// Monitors returns every output known to Hyprland, including disabled ones.
// It returns an empty slice without error when no Hyprland session is running,
// so callers can degrade gracefully on other desktops.
func Monitors() ([]Monitor, error) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		return nil, nil
	}
	if !CommandExists("hyprctl") {
		return nil, nil
	}
	out, err := exec.Command("hyprctl", "monitors", "all", "-j").Output()
	if err != nil {
		return nil, fmt.Errorf("query monitors: %w", err)
	}
	return ParseMonitors(out)
}

// ParseMonitors decodes the JSON emitted by `hyprctl monitors all -j`.
func ParseMonitors(data []byte) ([]Monitor, error) {
	var monitors []Monitor
	if err := json.Unmarshal(data, &monitors); err != nil {
		return nil, fmt.Errorf("parse monitors: %w", err)
	}
	return monitors, nil
}

// Rotated reports whether the output is rotated by 90 or 270 degrees.
func (m Monitor) Rotated() bool {
	return m.Transform == 1 || m.Transform == 3 || m.Transform == 5 || m.Transform == 7
}

// Portrait reports whether the output is taller than it is wide on screen.
// A rotated landscape panel becomes portrait; a native portrait panel is
// already taller before any transform.
func (m Monitor) Portrait() bool {
	w, h := m.PhysicalSize()
	return h > w
}

// PhysicalSize returns the panel dimensions rotated into their on-screen
// orientation, before scaling.
func (m Monitor) PhysicalSize() (int, int) {
	if m.Rotated() {
		return m.Height, m.Width
	}
	return m.Width, m.Height
}

// LogicalSize returns the effective (scaled, orientation-applied) dimensions
// used for layout and pointer coordinates. It mirrors omarchy-monitor-state.
func (m Monitor) LogicalSize() (int, int) {
	w, h := m.PhysicalSize()
	scale := m.Scale
	if scale <= 0 {
		scale = 1
	}
	return int(math.Round(float64(w) / scale)), int(math.Round(float64(h) / scale))
}
