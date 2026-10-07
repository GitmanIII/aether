package extraction

import (
	"testing"

	"aether/internal/color"
)

func TestEqualWeightPixelsTruncatesToSmallest(t *testing.T) {
	a := []color.RGB{{R: 1}, {R: 1}, {R: 1}, {R: 1}}
	b := []color.RGB{{B: 2}, {B: 2}}

	got := equalWeightPixels([][]color.RGB{a, b})
	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}
	for i, px := range got[:2] {
		if px != a[i] {
			t.Errorf("pixel %d = %+v, want from the first image", i, px)
		}
	}
	for i, px := range got[2:] {
		if px != b[i] {
			t.Errorf("pixel %d = %+v, want from the second image", i+2, px)
		}
	}
}

func TestEqualWeightPixelsSingleImageUnchanged(t *testing.T) {
	a := []color.RGB{{R: 5}, {G: 6}}
	got := equalWeightPixels([][]color.RGB{a})
	if len(got) != 2 || got[0] != a[0] || got[1] != a[1] {
		t.Fatalf("single image changed: %+v", got)
	}
}

func TestEqualWeightPixelsEmpty(t *testing.T) {
	if got := equalWeightPixels(nil); got != nil {
		t.Fatalf("nil samples = %+v, want nil", got)
	}
}
