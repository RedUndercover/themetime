package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"time"

	"fyne.io/systray"
)

type trayController struct {
	app        *App
	activeItem *systray.MenuItem
	nextItem   *systray.MenuItem
}

func newTrayController(app *App) *trayController {
	return &trayController{app: app}
}

func (t *trayController) externalLoop() (start, stop func()) {
	return systray.RunWithExternalLoop(t.ready, func() {})
}

func (t *trayController) ready() {
	systray.SetIcon(themeTimeIconPNG())
	systray.SetTitle("ThemeTime")
	systray.SetTooltip("ThemeTime solar theme scheduler")
	systray.SetOnTapped(t.app.showWindow)

	showItem := systray.AddMenuItem("Show ThemeTime", "Open the ThemeTime window")
	t.activeItem = systray.AddMenuItem("Active: Loading…", "Currently active theme rule")
	t.activeItem.Disable()
	t.nextItem = systray.AddMenuItem("Next: Loading…", "Next scheduled theme rule")
	t.nextItem.Disable()
	systray.AddSeparator()
	applyItem := systray.AddMenuItem("Apply current phase", "Apply the rule active at this time")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit", "Quit ThemeTime")

	go func() {
		for range showItem.ClickedCh {
			t.app.showWindow()
		}
	}()
	go func() {
		for range applyItem.ClickedCh {
			t.applyCurrent()
		}
	}()
	go func() {
		for range quitItem.ClickedCh {
			t.app.quit()
			return
		}
	}()
	go func() {
		for range systray.TrayOpenedCh {
			t.refreshStatus()
		}
	}()

	t.refreshStatus()
}

func (t *trayController) refreshStatus() {
	active, next, err := t.app.scheduleSummary()
	if err != nil {
		log.Printf("ThemeTime tray: refresh schedule: %v", err)
		if t.activeItem != nil {
			t.activeItem.SetTitle("Active: Unavailable")
		}
		if t.nextItem != nil {
			t.nextItem.SetTitle("Next: Unavailable")
		}
		return
	}
	if active == "" {
		active = "None"
	}
	if next == "" {
		next = "None"
	}
	if t.activeItem != nil {
		t.activeItem.SetTitle("Active: " + active)
	}
	if t.nextItem != nil {
		t.nextItem.SetTitle("Next: " + next)
	}
	systray.SetTooltip(fmt.Sprintf("ThemeTime · Active: %s", active))
}

func (t *trayController) applyCurrent() {
	phase, results, err := t.app.applyCurrent()
	if err == nil {
		for _, result := range results {
			if result.Error != "" {
				err = fmt.Errorf("%s", result.Error)
				break
			}
		}
	}
	if err != nil {
		log.Printf("ThemeTime tray: apply current phase: %v", err)
		systray.SetTooltip("ThemeTime · Apply failed: " + err.Error())
	} else {
		systray.SetTooltip("ThemeTime · Applied " + phase)
	}
	time.AfterFunc(4*time.Second, t.refreshStatus)
}

func themeTimeIconPNG() []byte {
	const size = 64
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	night := color.NRGBA{R: 24, G: 43, B: 83, A: 255}
	dusk := color.NRGBA{R: 210, G: 111, B: 91, A: 255}
	deep := color.NRGBA{R: 13, G: 35, B: 55, A: 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRoundedSquare(x, y, size, 13) {
				continue
			}
			horizontal := float64(x) / float64(size-1)
			vertical := float64(y) / float64(size-1)
			base := blend(night, dusk, math.Max(0, math.Min(1, (horizontal-0.22)*1.2)))
			base = blend(base, deep, math.Max(0, (vertical-0.58)*1.9))
			img.SetNRGBA(x, y, base)
		}
	}

	// Mirror the observatory silhouette in the launcher SVG so the native
	// window, system tray, and Plasma launcher read as the same icon.
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRoundedSquare(x, y, size, 13) || float64(y) < observatorySkyline(x) {
				continue
			}
			img.SetNRGBA(x, y, blend(img.NRGBAAt(x, y), color.NRGBA{R: 17, G: 43, B: 66, A: 255}, 0.54))
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - 42)
			dy := float64(y - 20)
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance <= 7 {
				img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 224, B: 122, A: 255})
			} else if distance <= 12 {
				amount := (12 - distance) / 5 * 0.42
				img.SetNRGBA(x, y, blend(img.NRGBAAt(x, y), color.NRGBA{R: 255, G: 224, B: 122, A: 255}, amount))
			}
		}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !insideRoundedSquare(x, y, size, 13) || !roundedSquareEdge(x, y, size, 13) {
				continue
			}
			img.SetNRGBA(x, y, blend(img.NRGBAAt(x, y), color.NRGBA{R: 255, G: 255, B: 255, A: 255}, 0.2))
		}
	}

	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		return nil
	}
	return output.Bytes()
}

func insideRoundedSquare(x, y, size, radius int) bool {
	if x < 0 || y < 0 || x >= size || y >= size {
		return false
	}
	if x >= radius && x < size-radius || y >= radius && y < size-radius {
		return true
	}
	cx := radius
	if x >= size-radius {
		cx = size - radius - 1
	}
	cy := radius
	if y >= size-radius {
		cy = size - radius - 1
	}
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func roundedSquareEdge(x, y, size, radius int) bool {
	return !insideRoundedSquare(x-1, y, size, radius) ||
		!insideRoundedSquare(x+1, y, size, radius) ||
		!insideRoundedSquare(x, y-1, size, radius) ||
		!insideRoundedSquare(x, y+1, size, radius)
}

func observatorySkyline(x int) float64 {
	points := [...]struct{ x, y float64 }{
		{-4, 54}, {12, 44}, {24, 49}, {36, 38}, {49, 45}, {68, 33},
	}
	px := float64(x)
	for index := 1; index < len(points); index++ {
		left, right := points[index-1], points[index]
		if px <= right.x {
			amount := (px - left.x) / (right.x - left.x)
			return left.y + (right.y-left.y)*amount
		}
	}
	return points[len(points)-1].y
}

func blend(a, b color.NRGBA, amount float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(a.R)*(1-amount) + float64(b.R)*amount),
		G: uint8(float64(a.G)*(1-amount) + float64(b.G)*amount),
		B: uint8(float64(a.B)*(1-amount) + float64(b.B)*amount),
		A: 255,
	}
}
