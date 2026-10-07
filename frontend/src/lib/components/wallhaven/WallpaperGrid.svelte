<script lang="ts">
    import WallpaperCard from './WallpaperCard.svelte';
    import ImagePreview from '$lib/components/shared/ImagePreview.svelte';
    import {getCardSize, CARD_MIN_WIDTH} from '$lib/stores/cardsize.svelte';
    import {getAssignTarget} from '$lib/stores/wallhaven.svelte';
    import {
        getDisplays,
        getActiveDisplayKey,
    } from '$lib/stores/displays.svelte';
    import {getIsApplying} from '$lib/stores/theme.svelte';
    import type {wallhaven} from '../../../../wailsjs/go/models';

    type Wallpaper = wallhaven.WallpaperInfo;

    let {wallpapers}: {wallpapers: Wallpaper[]} = $props();
    let previewIndex = $state(-1);
    // Derive shared card state once for the whole grid instead of per card.
    let target = $derived(
        getAssignTarget()
            ? (getDisplays().find(d => d.key === getAssignTarget()) ?? null)
            : null
    );
    // Fall back to the editor's elected monitor so "Use" always lands on a
    // monitor instead of desyncing the hero from the per-display assignment.
    let elected = $derived(
        getActiveDisplayKey()
            ? (getDisplays().find(d => d.key === getActiveDisplayKey()) ?? null)
            : null
    );
    let applying = $derived(getIsApplying());

    function getPreviewSrc(wp: Wallpaper): string {
        return wp.path || wp.thumbs?.original || wp.thumbs?.large;
    }
</script>

<div
    class="grid gap-3"
    style:grid-template-columns="repeat(auto-fill, minmax({CARD_MIN_WIDTH[
        getCardSize()
    ]}px, 1fr))"
>
    {#each wallpapers as wp, i (wp.id)}
        <WallpaperCard
            wallpaper={wp}
            {target}
            {elected}
            {applying}
            onpreview={() => (previewIndex = i)}
        />
    {/each}
</div>

<ImagePreview
    src={previewIndex >= 0 ? getPreviewSrc(wallpapers[previewIndex]) : ''}
    alt={previewIndex >= 0 ? wallpapers[previewIndex].id : ''}
    open={previewIndex >= 0}
    onclose={() => (previewIndex = -1)}
    hasPrev={previewIndex > 0}
    hasNext={previewIndex < wallpapers.length - 1}
    onprev={() => previewIndex--}
    onnext={() => previewIndex++}
/>
