package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	trayIconOnce sync.Once
	trayIcon     fyne.Resource
)

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
	const size = 64
	const radius = 22.0
	const stroke = 6.0
	center := float64(size-1) / 2.0

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	white := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Hypot(dx, dy)
			if math.Abs(dist-radius) <= stroke/2.0 {
				img.Set(x, y, white)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return fyne.NewStaticResource("tray.png", buf.Bytes())
}
