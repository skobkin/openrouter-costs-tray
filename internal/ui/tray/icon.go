package tray

import (
	_ "embed"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	trayIconOnce sync.Once
	trayIcon     fyne.Resource
)

//go:embed assets/icon.png
var trayIconPNG []byte

func IconResource() fyne.Resource {
	trayIconOnce.Do(func() {
		trayIcon = buildTrayIcon()
		if trayIcon == nil {
			trayIcon = theme.InfoIcon()
		}
	})
	return trayIcon
}

func buildTrayIcon() fyne.Resource {
	if len(trayIconPNG) == 0 {
		return nil
	}
	return fyne.NewStaticResource("tray.png", trayIconPNG)
}
