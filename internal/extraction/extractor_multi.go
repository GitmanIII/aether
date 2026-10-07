package extraction

import (
	"fmt"

	"aether/internal/color"
	"aether/internal/theme"
)

// ExtractColorsFromImages blends multiple images into a single 16-color palette
// by sampling pixels from each image and concatenating them before quantization.
// Every image is weighted equally regardless of resolution: each contributes the
// same number of samples, so a 4K wallpaper cannot dominate a 1080p one. Non-image
// inputs and unreadable files are skipped; the second return value is the count of
// skipped paths, intended for UI feedback.
func ExtractColorsFromImages(imagePaths []string, lightMode bool, mode string) ([16]string, int, error) {
	if err := validateMode(mode); err != nil {
		return [16]string{}, 0, err
	}
	if len(imagePaths) == 0 {
		return [16]string{}, 0, fmt.Errorf("no images provided")
	}

	cacheKey := buildCacheKey(GetMultiCacheKey(imagePaths, lightMode), mode)
	if cacheKey != "" {
		if cached, ok := LoadCachedPalette(cacheKey); ok {
			return cached, 0, nil
		}
	}

	samples := make([][]color.RGB, 0, len(imagePaths))
	skipped := 0
	for _, p := range imagePaths {
		if p == "" || !theme.IsImageFile(p) {
			skipped++
			continue
		}
		px, err := LoadAndSamplePixels(p)
		if err != nil || len(px) == 0 {
			skipped++
			continue
		}
		samples = append(samples, px)
	}
	if len(samples) == 0 {
		return [16]string{}, skipped, fmt.Errorf("no readable images provided")
	}

	allPixels := equalWeightPixels(samples)

	dominantColors, counts, err := ExtractDominantColorsFromPixels(allPixels, DominantColorsToExtract)
	if err != nil {
		return [16]string{}, skipped, fmt.Errorf("color extraction failed: %w", err)
	}
	if len(dominantColors) < 8 {
		return [16]string{}, skipped, fmt.Errorf("not enough colors extracted from images")
	}

	weights := normalizeCounts(counts)
	palette := NormalizeBrightness(GeneratePaletteByMode(dominantColors, weights, lightMode, mode))

	if cacheKey != "" {
		SavePaletteToCache(cacheKey, palette)
	}
	return palette, skipped, nil
}

// equalWeightPixels concatenates per-image samples after truncating each to the
// smallest sample count, so every image contributes the same number of pixels
// (1/N) to a blended palette regardless of its resolution.
func equalWeightPixels(samples [][]color.RGB) []color.RGB {
	if len(samples) == 0 {
		return nil
	}
	limit := len(samples[0])
	for _, px := range samples[1:] {
		if len(px) < limit {
			limit = len(px)
		}
	}
	blended := make([]color.RGB, 0, limit*len(samples))
	for _, px := range samples {
		blended = append(blended, px[:limit]...)
	}
	return blended
}
