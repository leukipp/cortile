package ui

import (
	"bytes"
	"image"

	"image/color"
	"image/draw"
	"image/png"

	"fyne.io/systray"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

var (
	iconSize     int = 256 // Size of systray icon
	iconMargin   int = 36  // Margin of systray icon
	layoutMargin int = 12  // Margin of layout rectangles
)

func UpdateIcon(ws *desktop.Workspace) {
	location := store.Location{DeskNum: store.Workplace.CurrentDesk, ScreenNum: store.Workplace.CurrentScreen}
	if ws == nil || ws.Location != location || len(common.Config.TilingIcon) == 0 {
		return
	}

	// Obtain layout name
	name := ws.ActiveLayout().GetName()
	if ws.Disabled() {
		name = "disabled"
	}

	// Initialize image
	col := image.Uniform{rgba("icon_foreground")}
	icon := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))

	// Draw background rectangle
	x0, y0, x1, y1 := iconMargin, iconMargin, iconSize-iconMargin, iconSize-iconMargin
	draw.Draw(icon, icon.Bounds(), &image.Uniform{rgba("icon_background")}, image.Point{}, draw.Src)

	// Draw layout rectangles
	switch name {
	case "vertical-left":
		draw.Draw(icon, image.Rect(x0, y0, x0+(x1-x0)/2-layoutMargin, y1), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+(x1-x0)/2+layoutMargin, y0, x1, y0+(y1-y0)/2-layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+(x1-x0)/2+layoutMargin, y0+(y1-y0)/2+layoutMargin, x1, y1), &col, image.Point{}, draw.Src)
	case "vertical-right":
		draw.Draw(icon, image.Rect(x0, y0, x0+(x1-x0)/2-layoutMargin, y0+(y1-y0)/2-layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0, y0+(y1-y0)/2+layoutMargin, x0+(x1-x0)/2-layoutMargin, y1), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+(x1-x0)/2+layoutMargin, y0, x1, y1), &col, image.Point{}, draw.Src)
	case "horizontal-top":
		draw.Draw(icon, image.Rect(x0, y0, x1, y0+(y1-y0)/2-layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0, y0+(y1-y0)/2+layoutMargin, x0+(x1-x0)/2-layoutMargin, y1), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+(x1-x0)/2+layoutMargin, y0+(y1-y0)/2+layoutMargin, x1, y1), &col, image.Point{}, draw.Src)
	case "horizontal-bottom":
		draw.Draw(icon, image.Rect(x0, y0, x0+(x1-x0)/2-layoutMargin, y0+(y1-y0)/2-layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+(x1-x0)/2+layoutMargin, y0, x1, y0+(y1-y0)/2-layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0, y0+(y1-y0)/2+layoutMargin, x1, y1), &col, image.Point{}, draw.Src)
	case "maximized":
		draw.Draw(icon, image.Rect(x0, y0, x1, y0+(y1-y0)/5-layoutMargin/2), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0, y0+(y1-y0)/5+layoutMargin/2, x1, y1), &col, image.Point{}, draw.Src)
	case "fullscreen":
		draw.Draw(icon, image.Rect(x0, y0, x1, y1), &col, image.Point{}, draw.Src)
	case "disabled":
		draw.Draw(icon, image.Rect(x0, y0, x0+2*layoutMargin, y1-2*layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0, y0, x1-2*layoutMargin, y0+2*layoutMargin), &col, image.Point{}, draw.Src)
		draw.Draw(icon, image.Rect(x0+2*layoutMargin+20, y0+2*layoutMargin+20, x1, y1), &col, image.Point{}, draw.Src)
	}

	// Draw version checker
	latest := common.VersionToInt(common.Build.Latest)
	current := common.VersionToInt(common.Build.Version)
	if latest > current {
		col := image.Uniform{color.RGBA{
			R: uint8(250),
			G: uint8(80),
			B: uint8(30),
			A: uint8(255),
		}}
		dx, dy := iconSize/10, iconSize/10
		draw.Draw(icon, image.Rect(x1-dx, y1-dy, x1+dx, y1+dy), &col, image.Point{}, draw.Src)
	}

	// Encode image bytes
	data := new(bytes.Buffer)
	png.Encode(data, icon)

	// Update systray icon
	systray.SetIcon(data.Bytes())
}

func rgba(name string) color.RGBA {
	rgba := common.Config.Colors[name]

	// Validate length
	if len(rgba) != 4 {
		log.Warn("Error obtaining color for ", name)
		return color.RGBA{}
	}

	return color.RGBA{
		R: uint8(rgba[0]),
		G: uint8(rgba[1]),
		B: uint8(rgba[2]),
		A: uint8(rgba[3]),
	}
}
