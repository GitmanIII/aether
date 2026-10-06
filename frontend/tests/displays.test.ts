import {beforeEach, expect, test, vi} from 'vitest';
import DisplaysView from '../src/lib/components/displays/DisplaysView.svelte';
import * as displays from '../src/lib/stores/displays.svelte';
import {
    ListDisplays,
    SetDisplayWallpaper,
    ClearDisplayWallpaper,
} from '../wailsjs/go/main/App';
import type {omarchy} from '../wailsjs/go/models';
import {render, settle, button} from './setup';

vi.mock('../wailsjs/go/main/App', () => ({
    ListDisplays: vi.fn(),
    SetDisplayWallpaper: vi.fn().mockResolvedValue(undefined),
    ClearDisplayWallpaper: vi.fn().mockResolvedValue(undefined),
    GetThumbnail: vi.fn().mockResolvedValue('data:image/png;base64,AAAA'),
    OpenFileDialog: vi.fn(),
}));

function makeDisplay(
    overrides: Partial<omarchy.Display> = {}
): omarchy.Display {
    return {
        name: 'DP-4',
        description: 'LG Electronics LG ULTRAGEAR 010NTQDHH228',
        make: 'LG Electronics',
        model: 'LG ULTRAGEAR',
        serial: '010NTQDHH228',
        x: 2560,
        y: 0,
        width: 1440,
        height: 2560,
        physicalWidth: 1440,
        physicalHeight: 2560,
        scale: 1,
        transform: 3,
        portrait: true,
        focused: true,
        key: 'LG Electronics:LG ULTRAGEAR:010NTQDHH228',
        keys: ['LG Electronics:LG ULTRAGEAR:010NTQDHH228', 'DP-4'],
        ...overrides,
    } as unknown as omarchy.Display;
}

beforeEach(() => {
    vi.mocked(ListDisplays)
        .mockReset()
        .mockResolvedValue({perScreen: true, displays: [makeDisplay()]});
    vi.mocked(SetDisplayWallpaper).mockReset().mockResolvedValue(undefined);
    vi.mocked(ClearDisplayWallpaper).mockReset().mockResolvedValue(undefined);
});

test('loadDisplays stores the displays and per-screen capability', async () => {
    await displays.loadDisplays(true);
    expect(displays.getDisplays()).toHaveLength(1);
    expect(displays.getPerScreen()).toBe(true);
    expect(displays.getDisplays()[0].portrait).toBe(true);
});

test('assignDisplayWallpaper sends the stable key and reloads', async () => {
    await displays.loadDisplays(true);
    await displays.assignDisplayWallpaper(
        displays.getDisplays()[0],
        '/tmp/portrait.jpg'
    );
    expect(SetDisplayWallpaper).toHaveBeenCalledWith(
        ['LG Electronics:LG ULTRAGEAR:010NTQDHH228', 'DP-4'],
        '/tmp/portrait.jpg'
    );
    expect(ListDisplays).toHaveBeenCalledTimes(2);
});

test('clearDisplayWallpaper clears by key', async () => {
    await displays.loadDisplays(true);
    await displays.clearDisplayWallpaper(displays.getDisplays()[0]);
    expect(ClearDisplayWallpaper).toHaveBeenCalledWith([
        'LG Electronics:LG ULTRAGEAR:010NTQDHH228',
        'DP-4',
    ]);
});

test('the view labels portrait displays and marks the focused one', async () => {
    const {target} = render(DisplaysView, {});
    await settle();
    expect(target.textContent).toContain('Portrait');
    expect(target.textContent).toContain('Focused');
    expect(target.textContent).toContain('1440×2560');
});

test('per-display actions are disabled when no per-display service is active', async () => {
    vi.mocked(ListDisplays).mockResolvedValue({
        perScreen: false,
        displays: [makeDisplay()],
    });
    const {target} = render(DisplaysView, {});
    await settle();
    expect(target.textContent).toContain(
        "Per-display rendering isn't available"
    );
    expect(button(target, 'Choose wallpaper').disabled).toBe(true);
});

test('an assigned display offers theme generation from its wallpaper', async () => {
    vi.mocked(ListDisplays).mockResolvedValue({
        perScreen: true,
        displays: [
            makeDisplay({
                assignment: {
                    type: 'image',
                    path: '/tmp/portrait.jpg',
                },
            } as unknown as omarchy.DisplayAssignment),
        ],
    });
    const {target} = render(DisplaysView, {});
    await settle();
    expect(button(target, 'Replace wallpaper')).toBeTruthy();
    expect(button(target, 'Generate theme from this display')).toBeTruthy();
    expect(button(target, 'Clear')).toBeTruthy();
});
