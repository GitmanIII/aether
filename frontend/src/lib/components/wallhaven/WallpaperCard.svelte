<script lang="ts">
    import {
        setWallpaperPath,
        addAdditionalImage,
        getAdditionalImages,
    } from '$lib/stores/theme.svelte';
    import {setActiveTab, showToast} from '$lib/stores/ui.svelte';
    import {applyWallpaperOnly} from '$lib/actions/themeActions';
    import {
        setAssignTarget,
        setRatio,
        setExactResolution,
    } from '$lib/stores/wallhaven.svelte';
    import {
        assignDisplayWallpaper,
        monitorLabel,
        nextUnassignedDisplay,
        markSessionSet,
        setActiveDisplayKey,
    } from '$lib/stores/displays.svelte';
    import {nearestWallhavenRatio, resolutionQuery} from '$lib/utils/aspect';
    import {openURL} from '$lib/utils/browser';
    import {observeIntersection} from '$lib/utils/intersection';
    import {isFavorite, toggleFavorite} from '$lib/stores/favorites.svelte';
    import {formatFileSize} from '$lib/utils/format';
    import WallpaperTile from '$lib/components/shared/WallpaperTile.svelte';
    import type {omarchy} from '../../../../wailsjs/go/models';

    let {
        wallpaper,
        onpreview,
        target = null,
        elected = null,
        applying = false,
    }: {
        wallpaper: any;
        onpreview: () => void;
        target?: omarchy.Display | null;
        elected?: omarchy.Display | null;
        applying?: boolean;
    } = $props();
    // Explicit target wins; otherwise fall back to the editor's elected monitor.
    let assignTarget = $derived(target ?? elected);
    let isDownloading = $state(false);
    let favoriteKey = $derived(wallpaper.path || wallpaper.id);
    let isFavorited = $derived(isFavorite(favoriteKey));
    let cardEl = $state<HTMLDivElement | null>(null);
    let inView = $state(false);

    // Gate <img> on viewport so scrolled-past cards release decoded-image
    // memory.
    $effect(() => {
        if (!cardEl) return;
        return observeIntersection(
            cardEl,
            entry => {
                inView = entry.isIntersecting;
            },
            {rootMargin: '600px 0px'}
        );
    });

    async function handleUse() {
        isDownloading = true;
        try {
            const {DownloadWallpaper} = await import(
                '../../../../wailsjs/go/main/App'
            );
            const localPath = await DownloadWallpaper(wallpaper.path);
            if (assignTarget) {
                // Assign to the chosen display and stay in the browser, then
                // nudge the user to the next display that still needs one.
                setWallpaperPath(localPath);
                setActiveDisplayKey(assignTarget.key);
                await assignDisplayWallpaper(assignTarget, localPath, true);
                markSessionSet(assignTarget.key);
                const next = nextUnassignedDisplay();
                if (next) {
                    setAssignTarget(next.key);
                    setRatio(
                        nearestWallhavenRatio(
                            next.physicalWidth,
                            next.physicalHeight
                        )
                    );
                    setExactResolution(
                        resolutionQuery(next.physicalWidth, next.physicalHeight)
                    );
                    showToast(
                        `${monitorLabel(assignTarget)} set — now choose for ${monitorLabel(next)}`
                    );
                } else {
                    showToast(
                        `${monitorLabel(assignTarget)} set — all displays set`
                    );
                }
            } else {
                setWallpaperPath(localPath);
                setActiveTab('editor');
                showToast(
                    'Wallpaper selected — click Extract to generate palette'
                );
            }
        } catch (e: any) {
            showToast('Failed to download wallpaper');
        } finally {
            isDownloading = false;
        }
    }

    async function handleWallpaperOnly() {
        if (!assignTarget) {
            await applyWallpaperOnly(wallpaper.path);
            return;
        }
        try {
            const {DownloadWallpaper} = await import(
                '../../../../wailsjs/go/main/App'
            );
            const localPath = await DownloadWallpaper(wallpaper.path);
            setWallpaperPath(localPath);
            setActiveDisplayKey(assignTarget.key);
            await assignDisplayWallpaper(assignTarget, localPath, true);
            markSessionSet(assignTarget.key);
            showToast(
                `${monitorLabel(assignTarget)} wallpaper updated — palette unchanged`
            );
        } catch {
            showToast('Failed to apply wallpaper');
        }
    }

    async function handleFavorite() {
        try {
            await toggleFavorite(favoriteKey, 'wallhaven', {
                id: wallpaper.id,
                thumbUrl: wallpaper.thumbs?.small,
                resolution: wallpaper.resolution,
            });
        } catch (err) {
            console.error('ToggleFavorite failed', err);
            showToast('Could not update favorites');
        }
    }

    async function handleAddExtra() {
        try {
            showToast('Downloading wallpaper...');
            const {DownloadWallpaper} = await import(
                '../../../../wailsjs/go/main/App'
            );
            const localPath = await DownloadWallpaper(wallpaper.path);
            if (getAdditionalImages().includes(localPath)) {
                showToast('Already in additional images');
                return;
            }
            showToast(
                addAdditionalImage(localPath)
                    ? 'Added to additional images'
                    : 'Skipped: the theme already has a wallpaper with that filename'
            );
        } catch {
            showToast('Failed to download wallpaper');
        }
    }

    function handleVisit() {
        openURL(`https://wallhaven.cc/w/${wallpaper.id}`);
    }
</script>

<div bind:this={cardEl}>
    <WallpaperTile
        path={wallpaper.path}
        name={wallpaper.id}
        {isFavorited}
        busy={isDownloading}
        {applying}
        useLabel={isDownloading
            ? 'Loading…'
            : assignTarget
              ? `Set for ${monitorLabel(assignTarget)}`
              : 'Use'}
        useTitle={assignTarget
            ? `Download and set as ${monitorLabel(assignTarget)}'s wallpaper`
            : 'Download, set as wallpaper, and open in editor'}
        visitTitle="Open on wallhaven.cc"
        onuse={handleUse}
        onwallpaperonly={handleWallpaperOnly}
        {onpreview}
        onaddextra={handleAddExtra}
        onfavorite={handleFavorite}
        onvisit={handleVisit}
    >
        {#snippet thumb()}
            {#if inView}
                <img
                    src={wallpaper.thumbs?.large ||
                        wallpaper.thumbs?.original ||
                        wallpaper.thumbs?.small}
                    alt={wallpaper.id}
                    class="h-full w-full object-cover"
                    loading="lazy"
                />
            {/if}
        {/snippet}
        {#snippet meta()}
            <span class="text-fg-secondary flex-1 truncate font-mono"
                >{wallpaper.resolution?.replace('x', ' × ')}</span
            >
            <span class="text-fg-dimmed shrink-0"
                >{formatFileSize(wallpaper.file_size)}</span
            >
        {/snippet}
    </WallpaperTile>
</div>
