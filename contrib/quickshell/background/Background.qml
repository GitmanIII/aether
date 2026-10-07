import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import QtQuick
import QtQuick.Effects
import QtQuick.Shapes
import qs.Commons
import qs.Ui

Item {
  id: root

  readonly property string home: Quickshell.env("HOME")
  readonly property string stateHome: home + "/.local/state"
  readonly property string currentBackgroundLink: stateHome + "/omarchy/current/background"

  property string currentBackground: ""
  property string displayedBackground: ""
  property string incomingBackground: ""
  property string oldBackground: ""
  property bool finishingTransition: false
  property int backgroundVersion: 0
  property int revealStartedVersion: -1
  property int pendingThemeVersion: -1
  property string pendingColorsRaw: ""
  property string pendingShellRaw: ""
  property real revealProgress: 1

  // Per-display assignments. The stock service paints one global background on
  // every output; this map lets a display keep its own image. Keys are matched
  // serial-first (make:model:serial) then by connector name, so an assignment
  // survives connector renumbering.
  readonly property string assignmentsPath: home + "/.config/omarchy/backgrounds.json"
  property var configState: ({ version: 1, displays: {} })

  function imageUrl(path) {
    return Util.fileUrl(path)
  }

  function screenKeys(screen) {
    var connector = String(screen.name || "").trim()
    var make = String(screen.manufacturer || "").trim()
    var product = String(screen.model || "").trim()
    var serial = String(screen.serialNumber || "").trim()
    var keys = []
    if (serial) keys.push([make, product, serial].filter(Boolean).join(":"))
    if (connector) keys.push(connector)
    return keys
  }

  function assignmentForScreen(screen) {
    var displays = configState && configState.displays ? configState.displays : {}
    var keys = screenKeys(screen)
    for (var i = 0; i < keys.length; i++) {
      var assignment = displays[keys[i]]
      if (assignment && assignment.type === "image" && String(assignment.path || "").trim())
        return assignment
    }
    return null
  }

  function normalizeImageAssignment(value) {
    if (!value || typeof value !== "object") return null
    if (value.type === "image" && String(value.path || "").trim())
      return { type: "image", path: String(value.path).trim() }
    return null
  }

  function parseAssignments(raw) {
    try {
      var parsed = JSON.parse(String(raw || ""))
      if (!parsed || typeof parsed !== "object" || parsed.version !== 1)
        return { version: 1, displays: {} }
      var source = parsed.displays && typeof parsed.displays === "object" ? parsed.displays : {}
      var displays = {}
      Object.keys(source).forEach(function(key) {
        var assignment = normalizeImageAssignment(source[key])
        if (String(key).trim() && assignment) displays[String(key).trim()] = assignment
      })
      return { version: 1, displays: displays }
    } catch (error) {
      return { version: 1, displays: {} }
    }
  }

  function saveAssignments(next) {
    configState = next
    assignmentsFile.setText(JSON.stringify(next, null, 2) + "\n")
  }

  function setAssignment(screenKey, value) {
    var key = String(screenKey || "").trim()
    var assignment = normalizeImageAssignment(value)
    if (!key || !assignment) return false
    var next = parseAssignments(JSON.stringify(configState))
    next.displays[key] = assignment
    saveAssignments(next)
    return true
  }

  function clearAssignment(screenKey) {
    var key = String(screenKey || "").trim()
    if (!key || !configState.displays[key]) return false
    var next = parseAssignments(JSON.stringify(configState))
    delete next.displays[key]
    saveAssignments(next)
    return true
  }

  // Desktop global actions (the wallpaper selector and the theme switcher)
  // reset every monitor to the global background, so drop per-display
  // assignments there. Aether's own per-display assignments are set through
  // setForScreen and are not cleared here.
  function clearAssignments() {
    if (!configState || !configState.displays) return
    if (Object.keys(configState.displays).length === 0) return
    saveAssignments({ version: 1, displays: {} })
  }

  function refreshBackground() {
    if (!readlinkProc.running) readlinkProc.running = true
  }

  function setBackground(path, instant) {
    transitionBackground("", path, path, instant, false)
  }

  function transitionBackground(fromPath, path, finalPath, instant, force) {
    path = String(path || "").trim()
    finalPath = String(finalPath || path).trim()
    fromPath = String(fromPath || "").trim()
    if (!path || (!force && finalPath === currentBackground)) return
    currentBackground = finalPath
    backgroundVersion += 1
    revealStartedVersion = -1

    revealAnimation.stop()
    finishingTransition = false

    if (instant || !displayedBackground) {
      oldBackground = ""
      incomingBackground = ""
      displayedBackground = path
      revealProgress = 1
      return
    }

    oldBackground = fromPath || displayedBackground
    incomingBackground = path
    revealProgress = 0
  }

  function setPendingTheme(colorsB64, shellB64) {
    pendingColorsRaw = Util.decodeBase64(colorsB64)
    pendingShellRaw = Util.decodeBase64(shellB64)
    pendingThemeVersion = backgroundVersion
    pendingThemeFallbackTimer.restart()
  }

  function applyPendingTheme() {
    // Background polling can advance backgroundVersion while a theme switch is
    // pending; the latest theme payload should still apply.
    if (pendingThemeVersion < 0) return
    pendingThemeFallbackTimer.stop()
    Color.loadColors(pendingColorsRaw)
    // Color.loadShell also refreshes Style so the type scale flips with the
    // background reveal instead of waiting for a separate reload path.
    Color.loadShell(pendingShellRaw)
    Style.scheduleRefresh()
    pendingThemeVersion = -1
    pendingColorsRaw = ""
    pendingShellRaw = ""
  }

  function transitionBackgroundWithTheme(fromPath, path, finalPath, colorsB64, shellB64) {
    transitionBackground(fromPath, path, finalPath, false, true)
    setPendingTheme(colorsB64, shellB64)
    if (!incomingBackground || revealProgress >= 1) applyPendingTheme()
  }

  function startReveal(panel) {
    if (!incomingBackground) return
    panel.maskReady = true
    if (revealStartedVersion === backgroundVersion) return
    revealStartedVersion = backgroundVersion
    applyPendingTheme()
    revealAnimation.restart()
  }

  function openSelector() {
    if (!bgSwitchProc.running) bgSwitchProc.running = true
  }

  function openThemeSwitcher() {
    if (!themeSwitchProc.running) themeSwitchProc.running = true
  }

  Process {
    id: bgSwitchProc
    command: ["bash", "-c", "background=$(omarchy-theme-bg-switcher); if [[ -n $background ]]; then printf '%s' \"$background\"; fi"]
    stdout: StdioCollector {
      onStreamFinished: {
        var path = String(text || "").trim()
        if (!path) return
        root.clearAssignments()
        applyPickedProc.command = ["omarchy-theme-bg-set", path]
        applyPickedProc.running = true
      }
    }
  }

  Process {
    id: themeSwitchProc
    command: ["bash", "-c", "theme=$(omarchy-theme-switcher); if [[ -n $theme ]]; then printf '%s' \"$theme\"; fi"]
    stdout: StdioCollector {
      onStreamFinished: {
        var theme = String(text || "").trim()
        if (!theme) return
        root.clearAssignments()
        applyThemeProc.command = ["bash", "-c", "omarchy-theme-set \"$1\" >/dev/null 2>&1 &", "bash", theme]
        applyThemeProc.running = true
      }
    }
  }

  Process {
    id: applyPickedProc
    onExited: root.refreshBackground()
  }

  Process {
    id: applyThemeProc
    onExited: root.refreshBackground()
  }

  Process {
    id: readlinkProc
    command: ["readlink", "-f", root.currentBackgroundLink]
    stdout: StdioCollector {
      onStreamFinished: root.setBackground(String(text || "").trim(), false)
    }
  }

  FileView {
    id: assignmentsFile
    path: root.assignmentsPath
    watchChanges: true
    printErrors: false
    onLoaded: root.configState = root.parseAssignments(text())
    onLoadFailed: root.configState = { version: 1, displays: {} }
    onFileChanged: reload()
  }

  IpcHandler {
    target: "background"

    function refresh(): void {
      root.refreshBackground()
    }

    function set(path: string): void {
      root.setBackground(path, false)
    }

    function setInstant(path: string): void {
      root.setBackground(path, true)
    }

    function transition(fromPath: string, path: string): void {
      root.transitionBackground(fromPath, path, path, false, false)
    }

    function themeTransition(fromPath: string, path: string, finalPath: string, colorsB64: string, shellB64: string): void {
      root.transitionBackgroundWithTheme(fromPath, path, finalPath, colorsB64, shellB64)
    }

    function setForScreen(screenKey: string, path: string): string {
      return root.setAssignment(screenKey, { type: "image", path: path }) ? "ok" : "invalid"
    }

    function clearForScreen(screenKey: string): string {
      return root.clearAssignment(screenKey) ? "ok" : "missing"
    }

    function assignments(): string {
      return JSON.stringify(root.configState)
    }
  }

  Timer {
    id: pendingThemeFallbackTimer
    interval: 300
    repeat: false
    onTriggered: root.applyPendingTheme()
  }

  NumberAnimation {
    id: revealAnimation
    target: root
    property: "revealProgress"
    from: 0
    to: 1
    duration: 420
    easing.type: Easing.InOutCubic
    onFinished: {
      if (root.incomingBackground) {
        root.displayedBackground = root.currentBackground || root.incomingBackground
        root.finishingTransition = true
      }
      root.revealProgress = 1
    }
  }

  Component.onCompleted: refreshBackground()

  Variants {
    model: Quickshell.screens

    PanelWindow {
      id: panel
      required property var modelData

      // Per-display image; empty means this output uses the global background.
      property var assignment: root.assignmentForScreen(modelData)
      readonly property string assignedPath: assignment ? assignment.path : ""
      // If an assigned file is missing (e.g. it lived in a replaced theme dir),
      // fall back to the global background instead of a blank desktop.
      property bool assignedBroken: false
      onAssignedPathChanged: assignedBroken = false

      screen: modelData
      visible: !remapGuard.remapping
      anchors { top: true; bottom: true; left: true; right: true }

      ScreenMoveRemap {
        id: remapGuard
        window: panel
      }
      color: "transparent"
      // Keep render updates enabled. The background layer has been observed to
      // lose its committed buffer while parked with updatesEnabled=false,
      // leaving a black desktop until omarchy-shell is restarted. The wallpaper
      // itself is static, so this favors correctness over a small render-loop
      // optimization.
      updatesEnabled: true

      property bool maskReady: false

      function maybeStartReveal() {
        if (!root.incomingBackground || root.revealProgress !== 0 || maskReady) return
        if (incomingFrame.status !== Image.Ready) return
        Qt.callLater(function() {
          if (!root.incomingBackground || root.revealProgress !== 0 || maskReady) return
          if (incomingFrame.status !== Image.Ready) return
          root.startReveal(panel)
        })
      }

      WlrLayershell.namespace: "omarchy-background"
      WlrLayershell.layer: WlrLayer.Background
      WlrLayershell.keyboardFocus: WlrKeyboardFocus.None
      exclusionMode: ExclusionMode.Ignore

      Image {
        id: base
        anchors.fill: parent
        source: root.imageUrl(panel.assignedPath !== "" && !panel.assignedBroken
          ? panel.assignedPath : root.displayedBackground)
        fillMode: Image.PreserveAspectCrop
        asynchronous: true
        cache: true
        onStatusChanged: {
          if (panel.assignedPath !== "" && status === Image.Error) {
            panel.assignedBroken = true
          } else if (status === Image.Ready) {
            panel.assignedBroken = false
          }
          if (status === Image.Ready && root.finishingTransition) {
            root.incomingBackground = ""
            root.oldBackground = ""
            root.finishingTransition = false
          }
        }
      }

      Image {
        id: oldFrame
        anchors.fill: parent
        source: root.imageUrl(root.oldBackground)
        fillMode: Image.PreserveAspectCrop
        asynchronous: true
        cache: false
        smooth: true
        mipmap: true
        visible: panel.assignedPath === "" && root.oldBackground !== "" && root.revealProgress < 1
        onStatusChanged: panel.maybeStartReveal()
      }

      Item {
        id: incomingLayer
        anchors.fill: parent
        visible: panel.assignedPath === "" && root.incomingBackground !== "" && incomingFrame.status === Image.Ready && (root.revealProgress >= 1 || panel.maskReady)
        layer.enabled: panel.assignedPath === "" && root.incomingBackground !== "" && root.revealProgress < 1
        layer.smooth: true
        layer.effect: MultiEffect {
          maskEnabled: true
          maskSource: revealMask
          maskThresholdMin: 0.5
          maskSpreadAtMin: 0.02
        }

        Image {
          id: incomingFrame
          anchors.fill: parent
          source: root.imageUrl(root.incomingBackground)
          fillMode: Image.PreserveAspectCrop
          asynchronous: true
          cache: false
          smooth: true
          mipmap: true
          onStatusChanged: panel.maybeStartReveal()
        }
      }

      Item {
        id: revealMask
        anchors.fill: parent
        visible: false
        layer.enabled: true

        readonly property real slant: -0.18
        readonly property real centerTop: width / 2 - slant * height / 2
        readonly property real centerBottom: width / 2 + slant * height / 2
        readonly property real reach: width / 2 + Math.abs(slant) * height / 2 + 4
        readonly property real spread: reach * root.revealProgress

        Shape {
          anchors.fill: parent
          antialiasing: true
          preferredRendererType: Shape.CurveRenderer
          ShapePath {
            fillColor: "white"
            strokeColor: "transparent"
            startX: revealMask.centerTop - revealMask.spread; startY: 0
            PathLine { x: revealMask.centerTop + revealMask.spread; y: 0 }
            PathLine { x: revealMask.centerBottom + revealMask.spread; y: revealMask.height }
            PathLine { x: revealMask.centerBottom - revealMask.spread; y: revealMask.height }
            PathLine { x: revealMask.centerTop - revealMask.spread; y: 0 }
          }
        }
      }

      Connections {
        target: root
        function onIncomingBackgroundChanged() {
          panel.maskReady = false
          panel.maybeStartReveal()
        }
      }

      MouseArea {
        anchors.fill: parent
        acceptedButtons: Qt.LeftButton | Qt.RightButton
        onDoubleClicked: function(mouse) {
          if (mouse.button === Qt.RightButton) root.openThemeSwitcher()
          else root.openSelector()
          mouse.accepted = true
        }
      }
    }
  }
}
