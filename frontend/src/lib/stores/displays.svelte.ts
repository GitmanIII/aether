import type {omarchy} from '../../../wailsjs/go/models';
import {showToast} from '$lib/stores/ui.svelte';

// --- Reactive state ---
let displays = $state<omarchy.Display[]>([]);
let perScreen = $state<boolean>(false);
let loading = $state<boolean>(false);
let loaded = $state<boolean>(false);
let error = $state<string>('');

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
    path: string
): Promise<void> {
    try {
        const {SetDisplayWallpaper} = await import(
            '../../../wailsjs/go/main/App'
        );
        await SetDisplayWallpaper(display.key, path);
        showToast(`Wallpaper set for ${display.name}`);
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
        await ClearDisplayWallpaper(display.key);
        showToast(`${display.name} now uses the global background`);
        await loadDisplays(true);
    } catch (err) {
        console.error('ClearDisplayWallpaper failed', err);
        showToast('Couldn’t clear the display wallpaper');
    }
}
