import type {omarchy} from '../../../wailsjs/go/models';
import {showToast} from '$lib/stores/ui.svelte';

// --- Reactive state ---
let displays = $state<omarchy.Display[]>([]);
let perScreen = $state<boolean>(false);
let loading = $state<boolean>(false);
let loaded = $state<boolean>(false);
let error = $state<string>('');
// Display keys the user has assigned during this run. Resets on launch so the
// picker shows "set" only for wallpapers chosen in the current session.
let sessionSet = $state<string[]>([]);
// The display whose wallpaper the editor currently shows/edits.
let activeKey = $state<string>('');

// --- Getters ---
export function getDisplays(): omarchy.Display[] {
    return displays;
}
export function getPerScreen(): boolean {
    return perScreen;
}
export function getDisplaysLoading(): boolean {
    return loading;
}
export function getDisplaysLoaded(): boolean {
    return loaded;
}
export function getDisplaysError(): string {
    return error;
}

/** Displays that currently have an image assigned. */
export function getAssignedDisplays(): omarchy.Display[] {
    return displays.filter(
        display =>
            display.assignment?.type === 'image' && !!display.assignment.path
    );
}

/** Display keys assigned during this run. */
export function getSessionSet(): string[] {
    return sessionSet;
}

/** Records that a display was assigned in this run. */
export function markSessionSet(key: string): void {
    if (key && !sessionSet.includes(key)) {
        sessionSet = [...sessionSet, key];
    }
}

/** The display whose wallpaper the editor is showing. */
export function getActiveDisplayKey(): string {
    return activeKey;
}
export function setActiveDisplayKey(key: string): void {
    activeKey = key;
}

/** 1-based position of a display in the layout, or 0 when unknown. */
export function monitorNumber(display: omarchy.Display): number {
    const index = displays.findIndex(d => d.key === display.key);
    return index >= 0 ? index + 1 : 0;
}

/** "Monitor 1", falling back to the connector name when the order is unknown. */
export function monitorLabel(display: omarchy.Display): string {
    const number = monitorNumber(display);
    return number > 0 ? `Monitor ${number}` : display.name;
}

/** First display not yet assigned in this run, in layout order. */
export function nextUnassignedDisplay(): omarchy.Display | null {
    return displays.find(d => !sessionSet.includes(d.key)) ?? null;
}

// --- Actions ---

/** Reads the connected displays and their per-display assignments. */
export async function loadDisplays(force = false): Promise<void> {
    if (loading) return;
    if (loaded && !force) return;
    loading = true;
    error = '';
    try {
        const {ListDisplays} = await import('../../../wailsjs/go/main/App');
        const result = await ListDisplays();
        displays = result.displays ?? [];
        perScreen = !!result.perScreen;
        loaded = true;
    } catch (err) {
        console.error('ListDisplays failed', err);
        error = 'Couldn’t read the display configuration';
    } finally {
        loading = false;
    }
}

/** Assigns an image to one display and refreshes the cached state. */
export async function assignDisplayWallpaper(
    display: omarchy.Display,
    path: string,
    silent = false
): Promise<void> {
    try {
        const {SetDisplayWallpaper} = await import(
            '../../../wailsjs/go/main/App'
        );
        await SetDisplayWallpaper(display.keys, path);
        if (!silent) showToast(`Wallpaper set for ${monitorLabel(display)}`);
        await loadDisplays(true);
    } catch (err) {
        console.error('SetDisplayWallpaper failed', err);
        showToast(
            'Couldn’t set a per-display wallpaper — a per-display background service must be running'
        );
    }
}

/** Removes a display's assignment so it uses the global background again. */
export async function clearDisplayWallpaper(
    display: omarchy.Display
): Promise<void> {
    try {
        const {ClearDisplayWallpaper} = await import(
            '../../../wailsjs/go/main/App'
        );
        await ClearDisplayWallpaper(display.keys);
        showToast(`${monitorLabel(display)} now uses the global background`);
        await loadDisplays(true);
    } catch (err) {
        console.error('ClearDisplayWallpaper failed', err);
        showToast('Couldn’t clear the display wallpaper');
    }
}
