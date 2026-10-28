//go:build windows

package clipboard

import (
	"golang.design/x/clipboard"
)

type winManager struct{}

func init() {
	clipboard.Init()
	constructor = func() Clipboard {
		return &winManager{}
	}
}

func (m *winManager) Read() ([]byte, error) {
	return clipboard.Read(clipboard.FmtText), nil
}

func (m *winManager) Write(data []byte) error {
	clipboard.Write(clipboard.FmtText, data)
	return nil
}
