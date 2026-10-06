import {beforeEach, expect, test, vi} from 'vitest';
import {
    nearestWallhavenRatio,
    resolutionsForRatio,
} from '../src/lib/utils/aspect';
import * as wallhaven from '../src/lib/stores/wallhaven.svelte';

vi.mock('../wailsjs/go/main/App', () => ({
    GetWallhavenConfig: vi.fn().mockResolvedValue({}),
    SaveWallhavenConfig: vi.fn().mockResolvedValue(undefined),
    SetWallhavenAPIKey: vi.fn().mockResolvedValue(undefined),
    SearchWallhaven: vi.fn().mockResolvedValue({
        data: [],
        meta: {last_page: 0, total: 0},
    }),
}));

beforeEach(() => {
    wallhaven.setRatio('');
    wallhaven.setExactResolution('');
});

test('nearest ratio maps common monitors', () => {
    expect(nearestWallhavenRatio(3840, 2160)).toBe('16x9');
    expect(nearestWallhavenRatio(1440, 2560)).toBe('9x16');
    expect(nearestWallhavenRatio(2560, 1440)).toBe('16x9');
    expect(nearestWallhavenRatio(1920, 1200)).toBe('16x10');
    expect(nearestWallhavenRatio(3440, 1440)).toBe('21x9');
});

test('resolutionsForRatio keeps an exact monitor resolution', () => {
    expect(resolutionsForRatio('9x16', '1440x2560')).toContain('1440x2560');
    expect(resolutionsForRatio('9x16')).toContain('1080x1920');
    expect(resolutionsForRatio('')).toEqual([]);
});

test('search params carry the selected ratio and resolution', () => {
    wallhaven.setRatio('9x16');
    wallhaven.setExactResolution('1440x2560');
    expect(wallhaven.buildSearchParams()).toEqual(
        expect.objectContaining({ratios: '9x16', resolutions: '1440x2560'})
    );
});

test('changing the ratio clears the exact resolution', () => {
    wallhaven.setExactResolution('1440x2560');
    wallhaven.setRatio('16x9');
    expect(wallhaven.getExactResolution()).toBe('');
    expect(wallhaven.buildSearchParams().ratios).toBe('16x9');
    expect(wallhaven.buildSearchParams().resolutions).toBe('');
});
