<script lang="ts">
    import {onMount} from 'svelte';
    import {
        getDisplays,
        loadDisplays,
        getActiveDisplayKey,
        setActiveDisplayKey,
        clearDisplayWallpaper,
        monitorLabel,
        getPerScreen,
    } from '$lib/stores/displays.svelte';
    import {getWallpaperPath, setWallpaperPath} from '$lib/stores/theme.svelte';
    import {setAssignTarget} from '$lib/stores/wallhaven.svelte';
    import {setActiveTab} from '$lib/stores/ui.svelte';
    import type {omarchy} from '../../../../wailsjs/go/models';

    let displays = $derived(getDisplays());
    let activeKey = $derived(getActiveDisplayKey());
    let wallpaper = $derived(getWallpaperPath());

    onMount(() => {
        loadDisplays(true);
    });

    // Elect the first display that has a wallpaper (default) once displays load.
    // Never switch tabs from here — only elect an existing wallpaper.
    $effect(() => {
        const list = displays;
        if (!list.length) return;
        if (activeKey && list.some(d => d.key === activeKey)) return;
        const first = list.find(hasWallpaper);
        if (first) elect(first);
    });

    function hasWallpaper(display: omarchy.Display): boolean {
        return (
            display.assignment?.type === 'image' && !!display.assignment.path
        );
    }

    function elect(display: omarchy.Display) {
        if (!hasWallpaper(display)) {
            // Nothing to edit yet — send the user to pick one for this monitor.
            setAssignTarget(display.key);
            setActiveTab('wallhaven');
            return;
        }
        setActiveDisplayKey(display.key);
        setWallpaperPath(display.assignment!.path!);
    }

    async function clear(display: omarchy.Display, event: MouseEvent) {
        event.stopPropagation();
        await clearDisplayWallpaper(display);
    }

    let bounds = $derived.by(() => {
        const list = displays;
        if (!list.length) return {x: 0, y: 0, width: 1, height: 1};
        let minX = Infinity;
        let minY = Infinity;
        let maxX = -Infinity;
        let maxY = -Infinity;
        for (const d of list) {
            minX = Math.min(minX, d.x);
            minY = Math.min(minY, d.y);
            maxX = Math.max(maxX, d.x + d.width);
            maxY = Math.max(maxY, d.y + d.height);
        }
        return {
            x: minX,
            y: minY,
            width: Math.max(1, maxX - minX),
            height: Math.max(1, maxY - minY),
        };
    });

    function rectStyle(display: omarchy.Display): string {
        const b = bounds;
        const left = ((display.x - b.x) / b.width) * 100;
        const top = ((display.y - b.y) / b.height) * 100;
        const width = (display.width / b.width) * 100;
        const height = (display.height / b.height) * 100;
        return `left:${left}%;top:${top}%;width:${width}%;height:${height}%`;
    }
</script>

{#if displays.length && wallpaper && getPerScreen()}
    <div
        class="pointer-events-none absolute inset-0 flex items-center justify-center p-5"
    >
        <div
            class="relative max-h-full w-full max-w-[460px]"
            style="aspect-ratio: {bounds.width} / {bounds.height};"
        >
            {#each displays as display (display.key)}
                {@const assigned = hasWallpaper(display)}
                {@const active = activeKey === display.key}
                <div
                    class="group pointer-events-auto absolute cursor-pointer border transition-colors
                        {active
                        ? 'border-accent bg-accent/10'
                        : assigned
                          ? 'border-white/45 bg-black/15 hover:border-white/70'
                          : 'border-dashed border-white/40 bg-black/20 hover:border-white/70'}"
                    style={rectStyle(display)}
                    role="button"
                    tabindex="0"
                    aria-pressed={active}
                    aria-label="{monitorLabel(display)}: {assigned
                        ? 'edit this wallpaper'
                        : 'choose a wallpaper'}"
                    onclick={() => elect(display)}
                    onkeydown={event => {
                        if (event.key === 'Enter' || event.key === ' ') {
                            event.preventDefault();
                            elect(display);
                        }
                    }}
                >
                    <div
                        class="pointer-events-none absolute inset-x-0 bottom-0 flex items-center justify-between gap-1 bg-black/45 px-1.5 py-0.5"
                    >
                        <span
                            class="truncate text-[10px] font-medium text-white"
                            >{monitorLabel(display)}</span
                        >
                        {#if !assigned}
                            <span class="shrink-0 text-[9.5px] text-white/70"
                                >choose</span
                            >
                        {/if}
                    </div>
                    {#if assigned}
                        <button
                            type="button"
                            class="pointer-events-auto absolute right-0 top-0 flex h-4 w-4 items-center justify-center text-[12px] leading-none text-white/70 hover:text-white"
                            title="Clear {monitorLabel(display)}'s wallpaper"
                            aria-label="Clear {monitorLabel(
                                display
                            )}'s wallpaper"
                            onclick={event => clear(display, event)}
                        >
                            ×
                        </button>
                    {/if}
                </div>
            {/each}
        </div>
    </div>
{/if}
