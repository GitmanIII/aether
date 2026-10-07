# Aether Omarchy shell plugins

Native `omarchy-shell` plugins for Aether:

- **`aether.wallpapers`** and **`aether.blueprints`** — overlays for selecting
  wallpapers and Aether blueprints. They use Omarchy shell colors and call the
  headless `aether` CLI, so the Aether GUI does not need to be open.
- **`aether.background`** — a background **service** that replaces
  `omarchy.background` and adds per-display wallpapers. It keeps the stock
  global background, transition, theme payload, and desktop double-click, so
  with no assignments the behavior is identical to stock.

## Install

From the repository root:

```bash
make install-omarchy-plugins
```

This installs and enables all three plugins.

## Open

```bash
omarchy-shell shell toggle aether.wallpapers '{}'
omarchy-shell shell toggle aether.blueprints '{}'
```

## Validate

```bash
omarchy plugin validate contrib/quickshell/wallpapers
omarchy plugin validate contrib/quickshell/blueprints
omarchy plugin validate contrib/quickshell/background
qmllint -I /usr/share/omarchy/shell contrib/quickshell/wallpapers/*.qml
qmllint -I /usr/share/omarchy/shell contrib/quickshell/blueprints/*.qml
qmllint -I /usr/share/omarchy/shell contrib/quickshell/background/*.qml
```

See [the shell plugin guide](../../docs/quickshell.md) and
[Displays](../../docs/displays.md) for details.
