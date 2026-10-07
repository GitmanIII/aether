package platform

import (
	"bufio"
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WatchMonitorEvents calls onChange whenever Hyprland reports a monitor being
// added or removed, so callers can refresh the display list on hotplug. It is a
// no-op when no Hyprland session is available. The watcher runs until ctx ends.
func WatchMonitorEvents(ctx context.Context, onChange func()) {
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if signature == "" || runtimeDir == "" {
		return
	}
	socket := filepath.Join(runtimeDir, "hypr", signature, ".socket2.sock")
	go watchLoop(ctx, socket, onChange)
}

func watchLoop(ctx context.Context, socket string, onChange func()) {
	for ctx.Err() == nil {
		conn, err := net.Dial("unix", socket)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}
		go func() {
			<-ctx.Done()
			conn.Close()
		}()
		readMonitorEvents(ctx, conn, onChange)
		conn.Close()
	}
}

func readMonitorEvents(ctx context.Context, conn net.Conn, onChange func()) {
	scanner := bufio.NewScanner(conn)
	var last time.Time
	for scanner.Scan() {
		if ctx.Err() != nil {
			return
		}
		if !isMonitorEvent(scanner.Text()) {
			continue
		}
		// Coalesce a burst (e.g. remove + re-add) into a single refresh: the
		// caller reads the full current layout anyway.
		if time.Since(last) < 300*time.Millisecond {
			continue
		}
		last = time.Now()
		onChange()
	}
}

// isMonitorEvent reports whether a Hyprland event line is a monitor hotplug
// event. It matches monitoradded, monitoraddedv2, monitorremoved, etc.
func isMonitorEvent(line string) bool {
	return strings.HasPrefix(line, "monitoradded") ||
		strings.HasPrefix(line, "monitorremoved")
}
