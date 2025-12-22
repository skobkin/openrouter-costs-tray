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
	const stroke = 5.0
	center := float64(size-1) / 2.0

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	white := color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	// Draw "O"
	drawRing(img, white, 20, int(center), 14, stroke)

	// Draw "R": stem + bowl + leg
	drawRect(img, white, 36, 14, 5, 36)
	drawRing(img, white, 44, 24, 9, stroke)
	drawRect(img, white, 40, 32, 9, 5)
	drawLine(img, white, 40, 34, 52, 50, 4)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return fyne.NewStaticResource("tray.png", buf.Bytes())
}

func drawRing(img *image.RGBA, c color.RGBA, cx, cy int, radius int, stroke float64) {
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Hypot(dx, dy)
			if math.Abs(dist-float64(radius)) <= stroke/2.0 {
				img.Set(x, y, c)
			}
		}
	}
}

func drawRect(img *image.RGBA, c color.RGBA, x, y, w, h int) {
	for yy := y; yy < y+h; yy++ {
		if yy < 0 || yy >= img.Rect.Dy() {
			continue
		}
		for xx := x; xx < x+w; xx++ {
			if xx < 0 || xx >= img.Rect.Dx() {
				continue
			}
			img.Set(xx, yy, c)
		}
	}
}

func drawLine(img *image.RGBA, c color.RGBA, x0, y0, x1, y1, thickness int) {
	steps := int(math.Max(math.Abs(float64(x1-x0)), math.Abs(float64(y1-y0))))
	if steps == 0 {
		return
	}
	r := thickness / 2
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x0) + (float64(x1-x0) * t)))
		y := int(math.Round(float64(y0) + (float64(y1-y0) * t)))
		for yy := y - r; yy <= y+r; yy++ {
			if yy < 0 || yy >= img.Rect.Dy() {
				continue
			}
			for xx := x - r; xx <= x+r; xx++ {
				if xx < 0 || xx >= img.Rect.Dx() {
					continue
				}
				img.Set(xx, yy, c)
			}
		}
	}
}
