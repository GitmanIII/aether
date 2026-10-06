// Aspect-ratio helpers for matching wallhaven searches to a display.

export type RatioOption = {id: string; label: string};

// Wallhaven only accepts this fixed set of ratios in its `ratios` query param.
export const WALLHAVEN_RATIOS: RatioOption[] = [
    {id: '16x9', label: '16:9 landscape'},
    {id: '16x10', label: '16:10 landscape'},
    {id: '21x9', label: '21:9 ultrawide'},
    {id: '32x9', label: '32:9 super ultrawide'},
    {id: '48x9', label: '48:9'},
    {id: '9x16', label: '9:16 portrait'},
    {id: '10x16', label: '10:16 portrait'},
    {id: '9x21', label: '9:21 portrait ultrawide'},
    {id: '9x32', label: '9:32 portrait'},
    {id: '9x48', label: '9:48 portrait'},
];

// Common resolutions offered once a ratio is chosen.
export const RESOLUTIONS_BY_RATIO: Record<string, string[]> = {
    '16x9': ['1280x720', '1920x1080', '2560x1440', '3840x2160'],
    '16x10': ['1280x800', '1680x1050', '1920x1200', '2560x1600', '3840x2400'],
    '21x9': ['2560x1080', '3440x1440', '5120x2160'],
    '32x9': ['3840x1080', '5120x1440'],
    '48x9': ['3840x720', '7680x1440'],
    '9x16': ['720x1280', '1080x1920', '1440x2560', '2160x3840'],
    '10x16': ['800x1280', '1200x1920', '1600x2560', '2400x3840'],
    '9x21': ['1080x2560', '1440x3440', '2160x5120'],
    '9x32': ['1080x3840', '1440x5120'],
    '9x48': ['720x3840', '1440x7680'],
};

function ratioAspect(id: string): number {
    const [w, h] = id.split('x').map(Number);
    return w > 0 && h > 0 ? w / h : 1;
}

/** Picks the closest wallhaven ratio for a monitor's physical resolution. */
export function nearestWallhavenRatio(width: number, height: number): string {
    if (!(width > 0) || !(height > 0)) return '';
    const logAspect = Math.log(width / height);
    let best = WALLHAVEN_RATIOS[0].id;
    let bestDistance = Infinity;
    for (const ratio of WALLHAVEN_RATIOS) {
        const distance = Math.abs(logAspect - Math.log(ratioAspect(ratio.id)));
        if (distance < bestDistance) {
            bestDistance = distance;
            best = ratio.id;
        }
    }
    return best;
}

/** Resolutions for a ratio, including an exact monitor resolution if given. */
export function resolutionsForRatio(ratio: string, extra = ''): string[] {
    const list = RESOLUTIONS_BY_RATIO[ratio] ?? [];
    if (extra && !list.includes(extra)) return [extra, ...list];
    return list;
}

/** Formats a physical resolution as a wallhaven `resolutions` value. */
export function resolutionQuery(width: number, height: number): string {
    if (!(width > 0) || !(height > 0)) return '';
    return `${width}x${height}`;
}
