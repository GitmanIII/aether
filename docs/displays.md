# Multiple displays

Aether reads the monitor layout from Omarchy/Hyprland and can assign a wallpaper
to each connected display. This lets a portrait monitor use a native portrait
image instead of a landscape one cropped to fit, and lets the palette be drawn
from one display or a blend of several.

## Reading the layout

The layout comes from `hyprctl monitors all -j`, so it reflects the monitors as
Hyprland has configured them:

- output name (`DP-4`, `HDMI-A-2`, ...)
- position, logical size, and scale
- rotation/transform, from which Aether derives portrait vs. landscape
- manufacturer, model, and serial number

A rotated output (transform `1`, `3`, `5`, or `7`) is treated as portrait. The
editor draws the monitors to scale over the wallpaper, so the orientation is
obvious at a glance. A monitor is shown as **Monitor 1**, **Monitor 2**, … in
layout order.

## The per-display background service

Omarchy's stock background service paints a single image on every screen. To
render a different wallpaper per display, Aether ships its own background
service plugin under `contrib/quickshell/background`, installed with the other
Aether shell plugins:

```bash
make install-omarchy-plugins
```

It replaces `omarchy.background` (the same way the Backdrop plugin does) and
keeps the standard `background` IPC target while adding:

```text
background setForScreen <screen-key> <path>
background clearForScreen <screen-key>
background assignments
```

An output uses its assignment when it has one and the global background
otherwise, so with no assignments the behavior is identical to stock. It keeps
the stock reveal transition, theme payload, and desktop double-click. If an
assigned file is missing, the output falls back to the global background instead
of a blank desktop.

Any other service that implements the same IPC methods also works; Aether
detects the capability and falls back to the single global wallpaper when no
per-display service is running.

## Assigning wallpapers

Assignments are keyed serial-first (`make:model:serial`) with the connector name
as a fallback, so they survive connector renumbering. Aether sets and clears
every candidate key for a display.

In the **Wallhaven** browser, choose a display in the **Choose for display…**
picker; the aspect ratio and resolution filters follow that display, and each
card's button becomes **Set for Monitor N**. The picker shows which monitors are
already set and nudges you to the next one.

In the **Editor**, the monitor overlay sits over the wallpaper. Click a monitor
to elect it (the wallpaper tools — edit, blur, eyedropper, change — act on it),
click an empty monitor to jump to Wallhaven for it, or use the **×** to clear
it.

The **Palette source** panel chooses where the colors come from:

- **Individual** — extract from one monitor; selecting it extracts immediately.
- **Blended** — mix any number of monitors, each weighted equally (1/N)
  regardless of resolution.

## Compatibility

- The desktop **double-click** (wallpaper selector) and **theme switcher** are
  Omarchy's global actions. They clear per-display assignments so the chosen
  wallpaper or theme applies to every monitor, then you can assign per display
  again.
- The lock screen continues to use Omarchy's global background.
- Disabling or removing `aether.background` restores the stock
  `omarchy.background` renderer.

## Saving a per-display theme

Aether **blueprints** store the per-display wallpapers alongside the palette, so
saving a blueprint captures "Monitor 1 = A, Monitor 2 = B" and applying it
restores both. Use **Apply theme ▾ → Save as new blueprint…** (Ctrl+J) or the
Blueprints tab. Omarchy theme folders are left untouched, and a blueprint used on
a machine with no per-display service simply falls back to its single wallpaper.

## Troubleshooting

- **No monitors shown** — Aether only reads monitors inside a Hyprland session.
- **Per-display options are missing** — no per-screen background service is
  active. Install the Aether shell plugins (above) or Backdrop.
- **A wallpaper change is rejected** — run
  `omarchy-shell -q background assignments` to see the keys the service stores.
