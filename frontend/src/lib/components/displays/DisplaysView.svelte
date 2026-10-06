<script lang="ts">
    import {onMount} from 'svelte';
    import {
        getDisplays,
        getPerScreen,
        getDisplaysLoading,
        getDisplaysLoaded,
        getDisplaysError,
        loadDisplays,
        assignDisplayWallpaper,
        clearDisplayWallpaper,
    } from '$lib/stores/displays.svelte';
    import {setActiveTab, showToast} from '$lib/stores/ui.svelte';
    import {setWallpaperPath} from '$lib/stores/theme.svelte';
    import {extractColors} from '$lib/actions/themeActions';
    import DisplayThumbnail from './DisplayThumbnail.svelte';
    import type {omarchy} from '../../../../wailsjs/go/models';

    const BACKDROP_INSTALL =
        'omarchy plugin add https://github.com/lgse/backdrop.git --enable';

    let displays = $derived(getDisplays());
    let perScreen = $derived(getPerScreen());
    let loading = $derived(getDisplaysLoading());
    let loaded = $derived(getDisplaysLoaded());
    let error = $derived(getDisplaysError());
    let busyKey = $state('');

    onMount(() => {
        loadDisplays(true);
    });

    // Bounding box of the physical layout, used to draw the display map to scale.
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

    function resolution(display: omarchy.Display): string {
        return `${display.physicalWidth}×${display.physicalHeight}`;
    }

    function assignedImage(display: omarchy.Display): string {
        return display.assignment?.type === 'image'
            ? display.assignment.path || ''
            : '';
    }

    async function chooseWallpaper(display: omarchy.Display) {
        try {
            const {OpenFileDialog} = await import(
                '../../../../wailsjs/go/main/App'
            );
            const path = await OpenFileDialog();
            if (!path) return;
            busyKey = display.key;
            await assignDisplayWallpaper(display, path);
        } catch {
            showToast('Couldn’t open the wallpaper picker');
        } finally {
            busyKey = '';
        }
    }

    async function clearDisplay(display: omarchy.Display) {
        busyKey = display.key;
        try {
            await clearDisplayWallpaper(display);
        } finally {
            busyKey = '';
        }
    }

    async function useForTheme(display: omarchy.Display) {
        const path = assignedImage(display);
        if (!path) return;
        setWallpaperPath(path);
        setActiveTab('editor');
        await extractColors();
    }

    async function copyInstall() {
        try {
            await navigator.clipboard.writeText(BACKDROP_INSTALL);
            showToast('Install command copied');
        } catch {
            showToast(
                'Couldn’t copy — select the command and copy it manually'
            );
        }
    }
</script>

<div class="h-full overflow-y-auto">
    <div class="mx-auto max-w-[860px] px-8 py-11">
        <header class="mb-9 flex items-start justify-between gap-6">
            <div>
                <p
                    class="text-accent mb-2 text-[10px] font-semibold uppercase tracking-[0.18em]"
                >
                    Hardware
                </p>
                <h1
                    class="text-fg-primary text-[26px] font-semibold tracking-[-0.01em]"
                >
                    Displays
                </h1>
                <p class="text-fg-secondary mt-2 text-[13px] leading-relaxed">
                    Assign a wallpaper to each connected monitor, including
                    portrait displays, and generate a theme from any of them.
                    The layout is read from Hyprland.
                </p>
            </div>
            <button
                type="button"
                class="border-border text-fg-dimmed hover:border-border-focus hover:text-fg-primary mt-1 shrink-0 border px-3 py-1.5 text-[12px] transition-colors disabled:opacity-40"
                onclick={() => loadDisplays(true)}
                disabled={loading}
            >
                Refresh
            </button>
        </header>

        {#if !perScreen && loaded}
            <div
                class="border-border bg-bg-secondary mb-8 border p-[18px]"
                role="note"
            >
                <h2 class="text-fg-primary mb-1.5 text-[13px] font-semibold">
                    Per-display rendering isn't available
                </h2>
                <p class="text-fg-secondary text-[12.5px] leading-relaxed">
                    Omarchy's stock background service paints one image on every
                    screen. Aether can drive a per-display background service
                    that keeps the standard
                    <code class="text-fg-primary">background</code> interface
                    and adds <code class="text-fg-primary">setForScreen</code> —
                    for example the Backdrop plugin.
                </p>
                <div class="mt-3 flex flex-wrap items-center gap-2">
                    <code
                        class="bg-bg-primary border-border text-fg-primary min-w-0 flex-1 overflow-x-auto border px-2.5 py-1.5 text-[11.5px]"
                        >{BACKDROP_INSTALL}</code
                    >
                    <button
                        type="button"
                        class="border-border text-fg-dimmed hover:border-border-focus hover:text-fg-primary shrink-0 border px-3 py-1.5 text-[12px] transition-colors"
                        onclick={copyInstall}>Copy</button
                    >
                </div>
            </div>
        {/if}

        {#if error}
            <p class="text-destructive text-[13px]">{error}</p>
        {:else if loading && !loaded}
            <p class="text-fg-dimmed text-[13px]">Reading displays…</p>
        {:else if !displays.length}
            <div
                class="border-border text-fg-secondary border border-dashed px-6 py-12 text-center text-[13px]"
            >
                No displays detected. This view reads Hyprland's monitor
                configuration, which is only available inside a Hyprland
                session.
            </div>
        {:else}
            <section aria-labelledby="display-map-heading" class="mb-9">
                <h2
                    id="display-map-heading"
                    class="text-fg-dimmed mb-2.5 text-[10px] font-semibold uppercase tracking-[0.14em]"
                >
                    Layout
                </h2>
                <div
                    class="border-border bg-bg-secondary mx-auto w-full max-w-[560px] border p-4"
                >
                    <div
                        class="relative w-full"
                        style="aspect-ratio: {bounds.width} / {bounds.height};"
                    >
                        {#each displays as display (display.key)}
                            <div
                                class="border-border-focus bg-bg-primary absolute flex flex-col items-center justify-center gap-1 overflow-hidden border p-1 text-center"
                                style={rectStyle(display)}
                            >
                                <span
                                    class="text-fg-primary max-w-full truncate text-[11px] font-medium"
                                    >{display.name}</span
                                >
                                <span
                                    class="text-fg-dimmed text-[10px] uppercase tracking-[0.12em]"
                                    >{display.portrait
                                        ? 'Portrait'
                                        : 'Landscape'}</span
                                >
                            </div>
                        {/each}
                    </div>
                </div>
            </section>

            <section aria-labelledby="display-list-heading">
                <h2
                    id="display-list-heading"
                    class="text-fg-dimmed mb-2.5 text-[10px] font-semibold uppercase tracking-[0.14em]"
                >
                    Monitors
                </h2>
                <div class="flex flex-col gap-3">
                    {#each displays as display (display.key)}
                        <div
                            class="border-border bg-bg-secondary flex items-start gap-4 border p-[18px]"
                        >
                            <div
                                class="border-border shrink-0 overflow-hidden border bg-black"
                                style="width:{display.portrait
                                    ? '96px'
                                    : '160px'};aspect-ratio:{display.width} / {display.height};"
                            >
                                {#if assignedImage(display)}
                                    <DisplayThumbnail
                                        path={assignedImage(display)}
                                        alt="{display.name} wallpaper"
                                    />
                                {:else if display.assignment?.type === 'color' && display.assignment.color}
                                    <div
                                        class="h-full w-full"
                                        style="background:{display.assignment
                                            .color}"
                                    ></div>
                                {:else}
                                    <div
                                        class="text-fg-dimmed flex h-full w-full items-center justify-center px-2 text-center text-[10.5px]"
                                    >
                                        Global background
                                    </div>
                                {/if}
                            </div>

                            <div class="min-w-0 flex-1">
                                <div
                                    class="flex flex-wrap items-center gap-x-2 gap-y-1"
                                >
                                    <span
                                        class="text-fg-primary text-[14px] font-semibold"
                                        >{display.name}</span
                                    >
                                    {#if display.focused}
                                        <span
                                            class="text-accent text-[10px] font-semibold uppercase tracking-[0.14em]"
                                            >Focused</span
                                        >
                                    {/if}
                                    {#if display.portrait}
                                        <span
                                            class="bg-accent-muted text-accent px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-[0.1em]"
                                            >Portrait</span
                                        >
                                    {/if}
                                </div>
                                <p
                                    class="text-fg-dimmed mt-1 truncate text-[12px]"
                                >
                                    {[
                                        resolution(display),
                                        display.model ||
                                            display.description ||
                                            display.make,
                                        display.scale !== 1
                                            ? `${display.scale}× scale`
                                            : '',
                                    ]
                                        .filter(Boolean)
                                        .join(' · ')}
                                </p>
                                <p class="text-fg-secondary mt-1.5 text-[12px]">
                                    {assignedImage(display)
                                        ? `Wallpaper: ${assignedImage(display)
                                              .split('/')
                                              .pop()}`
                                        : 'Using the global Omarchy background'}
                                </p>

                                <div class="mt-3 flex flex-wrap gap-2">
                                    <button
                                        type="button"
                                        class="border-border text-fg-dimmed hover:border-border-focus hover:text-fg-primary border px-3 py-1.5 text-[12px] transition-colors disabled:opacity-40"
                                        onclick={() => chooseWallpaper(display)}
                                        disabled={!perScreen ||
                                            busyKey === display.key}
                                    >
                                        {assignedImage(display)
                                            ? 'Replace wallpaper'
                                            : 'Choose wallpaper'}
                                    </button>
                                    {#if assignedImage(display)}
                                        <button
                                            type="button"
                                            class="border-border text-fg-dimmed hover:border-border-focus hover:text-fg-primary border px-3 py-1.5 text-[12px] transition-colors disabled:opacity-40"
                                            onclick={() => useForTheme(display)}
                                        >
                                            Generate theme from this display
                                        </button>
                                        <button
                                            type="button"
                                            class="border-border text-fg-dimmed hover:border-border-focus hover:text-fg-primary border px-3 py-1.5 text-[12px] transition-colors disabled:opacity-40"
                                            onclick={() =>
                                                clearDisplay(display)}
                                            disabled={busyKey === display.key}
                                        >
                                            Clear
                                        </button>
                                    {/if}
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            </section>
        {/if}
    </div>
</div>
