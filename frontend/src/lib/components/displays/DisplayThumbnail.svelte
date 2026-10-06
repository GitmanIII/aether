<script lang="ts">
    let {path, alt = ''}: {path: string; alt?: string} = $props();

    let url = $state('');
    let failed = $state(false);

    $effect(() => {
        const current = path;
        url = '';
        failed = false;
        if (!current) return;
        let cancelled = false;
        (async () => {
            try {
                const {GetThumbnail} = await import(
                    '../../../../wailsjs/go/main/App'
                );
                const data = await GetThumbnail(current);
                if (!cancelled) url = data;
            } catch {
                if (!cancelled) failed = true;
            }
        })();
        return () => {
            cancelled = true;
        };
    });
</script>

{#if url}
    <img src={url} {alt} class="h-full w-full object-cover" />
{:else if failed}
    <div
        class="text-fg-dimmed flex h-full w-full items-center justify-center text-[11px]"
    >
        Preview unavailable
    </div>
{:else}
    <div
        class="text-fg-dimmed flex h-full w-full items-center justify-center text-[11px]"
    >
        Loading…
    </div>
{/if}
