import {beforeEach, expect, test, vi} from 'vitest';
import * as displays from '../src/lib/stores/displays.svelte';
import {
    ListDisplays,
    SetDisplayWallpaper,
    ClearDisplayWallpaper,
} from '../wailsjs/go/main/App';
import type {omarchy} from '../wailsjs/go/models';

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

test('getAssignedDisplays only returns displays with an image', async () => {
    vi.mocked(ListDisplays).mockResolvedValue({
        perScreen: true,
        displays: [
            makeDisplay(),
            makeDisplay({
                name: 'HDMI-A-2',
                key: 'HDMI-A-2',
                keys: ['HDMI-A-2'],
                assignment: {
                    type: 'image',
                    path: '/tmp/hdmi.jpg',
                } as unknown as omarchy.DisplayAssignment,
            }),
            makeDisplay({
                name: 'DP-5',
                key: 'DP-5',
                keys: ['DP-5'],
                assignment: {
                    type: 'color',
                    color: '#000000',
                } as unknown as omarchy.DisplayAssignment,
            }),
        ],
    });
    await displays.loadDisplays(true);
    const assigned = displays.getAssignedDisplays();
    expect(assigned).toHaveLength(1);
    expect(assigned[0].name).toBe('HDMI-A-2');
});

test('monitorLabel numbers displays in layout order', async () => {
    vi.mocked(ListDisplays).mockResolvedValue({
        perScreen: true,
        displays: [
            makeDisplay({
                name: 'HDMI-A-2',
                key: 'HDMI-A-2',
                keys: ['HDMI-A-2'],
            }),
            makeDisplay({name: 'DP-4', key: 'DP-4', keys: ['DP-4']}),
        ],
    });
    await displays.loadDisplays(true);
    expect(displays.monitorLabel(displays.getDisplays()[0])).toBe('Monitor 1');
    expect(displays.monitorLabel(displays.getDisplays()[1])).toBe('Monitor 2');
});

test('nextUnassignedDisplay follows the session set', async () => {
    vi.mocked(ListDisplays).mockResolvedValue({
        perScreen: true,
        displays: [
            makeDisplay({
                name: 'HDMI-A-2',
                key: 'HDMI-A-2',
                keys: ['HDMI-A-2'],
            }),
            makeDisplay({name: 'DP-4', key: 'DP-4', keys: ['DP-4']}),
        ],
    });
    await displays.loadDisplays(true);
    expect(displays.nextUnassignedDisplay()?.key).toBe('HDMI-A-2');
    displays.markSessionSet('HDMI-A-2');
    expect(displays.nextUnassignedDisplay()?.key).toBe('DP-4');
    displays.markSessionSet('DP-4');
    expect(displays.nextUnassignedDisplay()).toBeNull();
});
