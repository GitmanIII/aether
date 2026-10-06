package omarchy

import (
	"reflect"
	"testing"

	"aether/internal/platform"
)

func TestDisplayKeysPrefersSerial(t *testing.T) {
	monitor := platform.Monitor{
		Name:   "DP-4",
		Make:   "LG Electronics",
		Model:  "LG ULTRAGEAR",
		Serial: "010NTQDHH228",
	}
	want := []string{"LG Electronics:LG ULTRAGEAR:010NTQDHH228", "DP-4"}
	if got := displayKeys(monitor); !reflect.DeepEqual(got, want) {
		t.Errorf("displayKeys = %v, want %v", got, want)
	}
}

func TestDisplayKeysWithoutSerial(t *testing.T) {
	monitor := platform.Monitor{Name: "eDP-1", Make: "BOE", Model: "0x0905"}
	want := []string{"eDP-1"}
	if got := displayKeys(monitor); !reflect.DeepEqual(got, want) {
		t.Errorf("displayKeys = %v, want %v", got, want)
	}
}

func TestDisplayKeysSkipsEmptyParts(t *testing.T) {
	monitor := platform.Monitor{Name: "DP-1", Model: "Generic", Serial: "abc"}
	want := []string{"Generic:abc", "DP-1"}
	if got := displayKeys(monitor); !reflect.DeepEqual(got, want) {
		t.Errorf("displayKeys = %v, want %v", got, want)
	}
}

func TestParseAssignments(t *testing.T) {
	raw := []byte(`{
	  "version": 1,
	  "displays": {
	    "LG Electronics:LG ULTRAGEAR:010NTQDHH228": {"type": "image", "path": "/tmp/portrait.jpg"},
	    "DP-1": {"type": "color", "color": "#112233"},
	    "DP-2": {"type": "gradient", "colors": ["#000000", "#ffffff"], "angle": 45},
	    "DP-3": {"type": "image", "path": "   "},
	    "": {"type": "image", "path": "/tmp/skip.jpg"},
	    "DP-5": {"type": "unknown"}
	  }
	}`)
	assignments, err := parseAssignments(raw)
	if err != nil {
		t.Fatalf("parseAssignments: %v", err)
	}
	want := map[string]DisplayAssignment{
		"LG Electronics:LG ULTRAGEAR:010NTQDHH228": {Type: "image", Path: "/tmp/portrait.jpg"},
		"DP-1": {Type: "color", Color: "#112233"},
		"DP-2": {Type: "gradient"},
	}
	if !reflect.DeepEqual(assignments, want) {
		t.Errorf("assignments = %v, want %v", assignments, want)
	}
}

func TestParseAssignmentsRejectsMalformed(t *testing.T) {
	if _, err := parseAssignments([]byte("not json")); err == nil {
		t.Fatal("expected an error for malformed JSON")
	}
}

func TestNewDisplayMergesSerialAssignment(t *testing.T) {
	monitor := platform.Monitor{
		Name:      "DP-4",
		Make:      "LG Electronics",
		Model:     "LG ULTRAGEAR",
		Serial:    "010NTQDHH228",
		Width:     2560,
		Height:    1440,
		Scale:     1,
		Transform: 3,
		Focused:   true,
	}
	assignments := map[string]DisplayAssignment{
		"LG Electronics:LG ULTRAGEAR:010NTQDHH228": {Type: "image", Path: "/tmp/portrait.jpg"},
	}

	display := newDisplay(monitor, assignments)
	if display.Assignment == nil {
		t.Fatal("expected an assignment")
	}
	if display.Assignment.Path != "/tmp/portrait.jpg" {
		t.Errorf("assignment path = %q", display.Assignment.Path)
	}
	if !display.Portrait {
		t.Error("display should be portrait")
	}
	if display.Width != 1440 || display.Height != 2560 {
		t.Errorf("display size = %dx%d, want 1440x2560", display.Width, display.Height)
	}
	if display.Key != "LG Electronics:LG ULTRAGEAR:010NTQDHH228" {
		t.Errorf("display key = %q", display.Key)
	}
}

func TestNewDisplayFallsBackToConnectorAssignment(t *testing.T) {
	monitor := platform.Monitor{Name: "DP-1", Width: 1920, Height: 1080, Scale: 1}
	assignments := map[string]DisplayAssignment{
		"DP-1": {Type: "color", Color: "#000000"},
	}
	display := newDisplay(monitor, assignments)
	if display.Assignment == nil || display.Assignment.Type != "color" {
		t.Fatalf("expected the connector assignment, got %+v", display.Assignment)
	}
}
