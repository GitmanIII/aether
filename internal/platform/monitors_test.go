package platform

import "testing"

const monitorsJSON = `[
  {
    "id": 0,
    "name": "HDMI-A-2",
    "description": "LG Electronics LG ULTRAGEAR+ 502NTNH45639",
    "make": "LG Electronics",
    "model": "LG ULTRAGEAR+",
    "serial": "502NTNH45639",
    "width": 3840,
    "height": 2160,
    "refreshRate": 59.997,
    "x": 0,
    "y": 0,
    "scale": 1.5,
    "transform": 0,
    "focused": false,
    "disabled": false
  },
  {
    "id": 1,
    "name": "DP-4",
    "description": "LG Electronics LG ULTRAGEAR 010NTQDHH228",
    "make": "LG Electronics",
    "model": "LG ULTRAGEAR",
    "serial": "010NTQDHH228",
    "width": 2560,
    "height": 1440,
    "refreshRate": 143.973,
    "x": 2560,
    "y": 0,
    "scale": 1,
    "transform": 3,
    "focused": true,
    "disabled": false
  },
  {
    "id": 2,
    "name": "DP-9",
    "width": 1920,
    "height": 1080,
    "scale": 1,
    "transform": 0,
    "disabled": true
  }
]`

func TestParseMonitors(t *testing.T) {
	monitors, err := ParseMonitors([]byte(monitorsJSON))
	if err != nil {
		t.Fatalf("ParseMonitors: %v", err)
	}
	if len(monitors) != 3 {
		t.Fatalf("monitors = %d, want 3", len(monitors))
	}

	hdmi := monitors[0]
	if hdmi.Rotated() || hdmi.Portrait() {
		t.Errorf("HDMI should be landscape: rotated=%v portrait=%v", hdmi.Rotated(), hdmi.Portrait())
	}
	if w, h := hdmi.LogicalSize(); w != 2560 || h != 1440 {
		t.Errorf("HDMI logical size = %dx%d, want 2560x1440", w, h)
	}

	dp := monitors[1]
	if !dp.Rotated() || !dp.Portrait() {
		t.Errorf("DP-4 should be portrait: rotated=%v portrait=%v", dp.Rotated(), dp.Portrait())
	}
	if w, h := dp.PhysicalSize(); w != 1440 || h != 2560 {
		t.Errorf("DP-4 physical size = %dx%d, want 1440x2560", w, h)
	}
	if w, h := dp.LogicalSize(); w != 1440 || h != 2560 {
		t.Errorf("DP-4 logical size = %dx%d, want 1440x2560", w, h)
	}
	if !dp.Focused {
		t.Error("DP-4 should be focused")
	}
}

func TestParseMonitorsRejectsMalformed(t *testing.T) {
	if _, err := ParseMonitors([]byte("{")); err == nil {
		t.Fatal("expected an error for malformed JSON")
	}
}

func TestLogicalSizeDefaultsScale(t *testing.T) {
	m := Monitor{Width: 3840, Height: 2160}
	if w, h := m.LogicalSize(); w != 3840 || h != 2160 {
		t.Errorf("logical size = %dx%d, want 3840x2160", w, h)
	}
}

func TestPortraitNativePanel(t *testing.T) {
	m := Monitor{Width: 1080, Height: 1920}
	if !m.Portrait() {
		t.Error("a native portrait panel should report portrait")
	}
	if m.Rotated() {
		t.Error("an unrotated native portrait panel should not report rotated")
	}
}
