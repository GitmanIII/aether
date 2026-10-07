package platform

import "testing"

func TestIsMonitorEvent(t *testing.T) {
	for _, line := range []string{
		"monitoradded>>DP-4",
		"monitoraddedv2>>1,DP-4,LG ULTRAGEAR",
		"monitorremoved>>DP-4",
		"monitorremovedv2>>1",
	} {
		if !isMonitorEvent(line) {
			t.Errorf("isMonitorEvent(%q) = false, want true", line)
		}
	}
	for _, line := range []string{
		"",
		"workspace>>1",
		"activewindow>>kitty",
		"openwindow>>abc,1,kitty",
	} {
		if isMonitorEvent(line) {
			t.Errorf("isMonitorEvent(%q) = true, want false", line)
		}
	}
}
