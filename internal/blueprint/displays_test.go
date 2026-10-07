package blueprint

import (
	"encoding/json"
	"testing"
)

func TestPaletteDataDisplaysRoundTrip(t *testing.T) {
	in := PaletteData{
		Colors:   []string{"#000000"},
		Displays: map[string]string{"DP-4": "/tmp/a.jpg", "HDMI-A-2": "/tmp/b.jpg"},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out PaletteData
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Displays) != 2 ||
		out.Displays["DP-4"] != "/tmp/a.jpg" ||
		out.Displays["HDMI-A-2"] != "/tmp/b.jpg" {
		t.Fatalf("displays = %v", out.Displays)
	}
}

func TestPaletteDataOmitsEmptyDisplays(t *testing.T) {
	data, err := json.Marshal(PaletteData{Colors: []string{"#000000"}})
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["displays"]; ok {
		t.Fatal("empty displays should be omitted from the blueprint JSON")
	}
}
