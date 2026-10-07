<script lang="ts">
    import {onMount} from 'svelte';
    import {
        getAssignedDisplays,
        getActiveDisplayKey,
        setActiveDisplayKey,
        loadDisplays,
        monitorLabel,
    } from '$lib/stores/displays.svelte';
    import {
        setWallpaperPath,
        setAdditionalImages,
    } from '$lib/stores/theme.svelte';
    import {extractColors} from '$lib/actions/themeActions';
    import DisplayThumbnail from '$lib/components/displays/DisplayThumbnail.svelte';
    import Segmented from '$lib/components/shared/Segmented.svelte';
    import type {omarchy} from '../../../../wailsjs/go/models';

    type SourceMode = 'individual' | 'blended';

    let assigned = $derived(getAssignedDisplays());
    let activeKey = $derived(getActiveDisplayKey());
    let mode = $state<SourceMode>('individual');
    let selected = $state<string[]>([]);

    const modes = [
        {
            value: 'individual',
            label: 'Individual',
            title: 'Extract the palette from one monitor',
        },
        {
            value: 'blended',
            label: 'Blended',
            title: 'Blend several monitors equally',
        },
    ] as const;

    onMount(() => {
        loadDisplays();
    });

    // Keep the blended selection in sync as assignments come and go.
    $effect(() => {
        const keys = assigned.map(d => d.key);
        const pruned = selected.filter(key => keys.includes(key));
        if (pruned.length !== selected.length) {
            selected = pruned;
        } else if (selected.length === 0 && keys.length) {
            selected = keys;
        }
    });

    async function elect(display: omarchy.Display) {
        const path = display.assignment?.path;
        if (!path) return;
        mode = 'individual';
        setActiveDisplayKey(display.key);
        setWallpaperPath(path);
        setAdditionalImages([]);
        await extractColors();
    }

    function toggle(display: omarchy.Display) {
        selected = selected.includes(display.key)
            ? selected.filter(key => key !== display.key)
            : [...selected, display.key];
        // A single selection is just an individual source, so switch modes
        // rather than leaving the blend action disabled.
        if (selected.length === 1) {
            const only = assigned.find(d => d.key === selected[0]);
            if (only?.assignment?.path) {
                mode = 'individual';
                setActiveDisplayKey(only.key);
                setWallpaperPath(only.assignment.path);
            }
        }
    }

    async function extractBlended() {
        const chosen = assigned.filter(
            d => selected.includes(d.key) && d.assignment?.path
        );
        if (chosen.length < 2) return;
        const paths = chosen.map(d => d.assignment!.path!);
        setWallpaperPath(paths[0]);
        setAdditionalImages(paths.slice(1));
        await extractColors({allImages: true});
    }
</script>

{#if assigned.length}
    <section
        class="border-border bg-bg-secondary border p-[18px]"
        aria-labelledby="display-sources-heading"
    >
        <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
            <div>
                <h2
                    id="display-sources-heading"
                    class="text-fg-dimmed text-[10px] font-semibold uppercase tracking-[0.14em]"
                >
                    Palette source
                </h2>
                <p class="text-fg-secondary mt-1 text-[12px]">
                    Select a monitor to extract its colors, or blend several
                    equally.
                </p>
            </div>
            <Segmented
                options={modes}
                value={mode}
                onchange={value => (mode = value as SourceMode)}
                label="Palette source"
            />
        </div>

        {#if mode === 'individual'}
            <div class="flex flex-wrap items-center gap-2">
                {#each assigned as display (display.key)}
                    <button
                        type="button"
                        class="border-border hover:border-border-focus flex items-center gap-2 border px-2 py-1.5 transition-colors
                            {activeKey === display.key ? 'border-accent' : ''}"
                        onclick={() => elect(display)}
                        aria-pressed={activeKey === display.key}
                        title="Use {monitorLabel(display)}'s wallpaper"
                    >
                        <span
                            class="border-border h-9 w-9 shrink-0 overflow-hidden border bg-black"
                        >
                            <DisplayThumbnail
                                path={display.assignment?.path || ''}
                                alt="{monitorLabel(display)} wallpaper"
                            />
                        </span>
                        <span class="text-left">
                            <span
                                class="text-fg-primary block text-[12px] font-medium"
                                >{monitorLabel(display)}</span
                            >
                            <span class="text-fg-dimmed block text-[10.5px]"
                                >{display.physicalWidth}×{display.physicalHeight}</span
                            >
                        </span>
                    </button>
                {/each}
            </div>
        {:else}
            <div class="flex flex-col gap-2">
                {#each assigned as display (display.key)}
                    <label
                        class="border-border flex cursor-pointer items-center gap-3 border px-3 py-2"
                    >
                        <input
                            type="checkbox"
                            class="accent-accent"
                            checked={selected.includes(display.key)}
                            onchange={() => toggle(display)}
                        />
                        <span
                            class="border-border h-8 w-8 shrink-0 overflow-hidden border bg-black"
                        >
                            <DisplayThumbnail
                                path={display.assignment?.path || ''}
                                alt="{monitorLabel(display)} wallpaper"
                            />
                        </span>
                        <span class="text-fg-primary text-[12px] font-medium"
                            >{monitorLabel(display)}</span
                        >
                        <span class="text-fg-dimmed text-[10.5px]"
                            >{display.physicalWidth}×{display.physicalHeight}</span
                        >
                    </label>
                {/each}
                <button
                    type="button"
                    class="border-border text-fg-primary hover:border-border-focus hover:text-accent self-start border px-3 py-1.5 text-[12px] transition-colors disabled:opacity-40"
                    onclick={extractBlended}
                    disabled={selected.length < 2}
                >
                    Blend {selected.length} displays equally
                </button>
            </div>
        {/if}
    </section>
{/if}
