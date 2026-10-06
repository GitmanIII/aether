# Multiple displays

Aether reads the monitor layout from Omarchy/Hyprland and assigns a wallpaper
to each connected display. This lets a portrait monitor use a native portrait
image instead of a landscape one cropped to fit.

Open the **Displays** tab to see the connected monitors, their physical layout,
orientation, and any wallpaper assigned to each one.

## Reading the layout

The layout comes from `hyprctl monitors all -j`, so it reflects the monitors as
Hyprland has configured them:

- output name (`DP-4`, `HDMI-A-2`, ...)
- position, logical size, and scale
- rotation/transform, from which Aether derives portrait vs. landscape
- manufacturer, model, and serial number

A rotated output (transform `1`, `3`, `5`, or `7`) is shown as portrait, and the
display map on the Displays tab is drawn to scale so the orientation is obvious.

## Per-display wallpapers

Omarchy's stock background service paints a single image on every screen. To
render a different wallpaper per display, Aether talks to a **per-display
background service** that keeps Omarchy's standard `background` IPC target and
adds:

```text
background setForScreen <screen-key> <path>
background clearForScreen <screen-key>
background assignments
```

[Backdrop](https://github.com/lgse/backdrop) is one such service:

```bash
omarchy plugin add https://github.com/lgse/backdrop.git --enable
```

When a per-display service is active, each monitor card on the Displays tab
offers:

- **Choose wallpaper** — pick an image for that display. Aether prefers the
  serial-backed key (`make:model:serial`) so the assignment survives connector
  renumbering, and falls back to the connector name.
- **Generate theme from this display** — uses that display's wallpaper as the
  color-extraction source and opens the editor.
- **Clear** — removes the assignment so the display uses Omarchy's global
  background again.

When no per-display service is running, the Displays tab still shows the
monitor layout and orientation, but per-display actions are disabled and Aether
applies the single global wallpaper as before.

## Compatibility

- Changing the background or applying a theme keeps each display's explicit
  assignment; it only updates the global fallback.
- The lock screen continues to use Omarchy's global background.
- Disabling or removing the per-display service restores the stock
  `omarchy.background` renderer.

## Troubleshooting

- **"No displays detected"** — Aether only reads monitors inside a Hyprland
  session. On other desktops the Displays tab has nothing to show.
- **The per-display banner is shown** — no per-display background service is
  active. Install one (see above) and use **Refresh**.
- **A wallpaper change is rejected** — the service may not recognize the
  display key. Run `omarchy-shell -q background assignments` to see the keys it
  stores.
