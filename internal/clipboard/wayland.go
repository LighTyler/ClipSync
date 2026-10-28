// go:build linux && wayland

package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type linuxManager struct{}

func init() {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		panic("clipboard: WAYLAND_DISPLAY not set (not running under Wayland)")
	}
	constructor = func() Clipboard {
		return &linuxManager{}
	}
}

func (m *linuxManager) Read() ([]byte, error) {
	// Вызываем wl-paste
	cmd := exec.Command("wl-paste", "--no-newline")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wl-paste failed: %w", err)
	}
	return out.Bytes(), nil
}

func (m *linuxManager) Write(data []byte) error {
	// Вызываем wl-copy
	cmd := exec.Command("wl-copy")
	cmd.Stdin = bytes.NewReader(data)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wl-copy failed: %w", err)
	}
	return nil
}
