package ui

import (
	"image"
	"image/draw"
	"math"
	"time"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/BurntSushi/freetype-go/freetype/truetype"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

var (
	fontSize   int = 16 // Size of text font
	fontMargin int = 4  // Margin of text font
	rectMargin int = 4  // Margin of layout rectangles
)

var (
	gui map[uint]*xwindow.Window = make(map[uint]*xwindow.Window) // Layout overlay window
)

func ShowLayout(ws *desktop.Workspace) {
	location := store.Location{DeskNum: store.CurrentDesk, ScreenNum: store.CurrentScreen}
	if ws == nil || ws.Location.DeskNum != location.DeskNum || common.Config.TilingGui <= 0 {
		return
	}

	// Wait for tiling events
	time.AfterFunc(150*time.Millisecond, func() {

		// Obtain layout name
		name := ws.ActiveLayout().GetName()
		if ws.Disabled() {
			name = "disabled"
		}

		// Calculate scaled desktop dimensions
		dim := xrect.New(store.DesktopDimensions(ws.Location.ScreenNum))
		_, _, width, height := scale(dim.X(), dim.Y(), dim.Width(), dim.Height())

		// Create an empty canvas image
		bg := bgra("gui_background")
		cv := xgraphics.New(store.X, image.Rect(0, 0, width+rectMargin, height+fontSize+2*fontMargin+2*rectMargin))
		cv.For(func(x int, y int) xgraphics.BGRA { return bg })

		// Draw client rectangles
		drawClients(cv, ws, name)

		// Draw layout name
		drawText(cv, name, bgra("gui_text"), cv.Rect.Dx()/2, cv.Rect.Dy()-2*fontMargin-rectMargin, fontSize)

		// Show the canvas graphics
		showGraphics(cv, ws, time.Duration(common.Config.TilingGui))
	})
}

func drawClients(cv *xgraphics.Image, ws *desktop.Workspace, layout string) {
	al := ws.ActiveLayout()
	mg := al.GetManager()

	// Obtain visible clients
	clients := mg.Clients(false)
	for _, c := range clients {
		if c == nil {
			return
		}
		for _, state := range c.Latest.States {
			if state == "_NET_WM_STATE_FULLSCREEN" || layout == "fullscreen" {
				clients = mg.Visible(&store.Clients{Items: mg.Clients(true), MaxAllowed: 1})
				break
			}
		}
	}

	// Draw default rectangle
	dim := xrect.New(store.DesktopDimensions(ws.Location.ScreenNum))
	if len(clients) == 0 || layout == "disabled" {

		// Calculate scaled desktop dimensions
		x, y, width, height := scale(0, 0, dim.Width(), dim.Height())

		// Draw client rectangle onto canvas
		color := bgra("gui_client_slave")
		drawImage(cv, &image.Uniform{color}, color, x+rectMargin, y+rectMargin, x+width, y+height)

		return
	}

	// Draw master and slave rectangle
	for _, c := range clients {

		// Calculate scaled client dimensions
		cx, cy, cw, ch := c.OuterGeometry()
		x, y, width, height := scale(cx-dim.X(), cy-dim.Y(), cw, ch)

		// Calculate icon size
		iconSize := math.MaxInt
		if width < iconSize {
			iconSize = width
		}
		if height < iconSize {
			iconSize = height
		}
		iconSize /= 2

		// Obtain rectangle color
		color := bgra("gui_client_slave")
		if mg.IsMaster(c) || layout == "fullscreen" {
			color = bgra("gui_client_master")
		}

		// Draw client rectangle onto canvas
		rect := &image.Uniform{color}
		drawImage(cv, rect, color, x+rectMargin, y+rectMargin, x+width, y+height)

		// Draw client icon onto canvas
		ico, err := xgraphics.FindIcon(store.X, c.Win.Id, iconSize, iconSize)
		if err == nil {
			drawImage(cv, ico, color, x+rectMargin/2+width/2-iconSize/2, y+rectMargin/2+height/2-iconSize/2, x+width, y+height)
		}
	}
}

func drawImage(cv *xgraphics.Image, img image.Image, color xgraphics.BGRA, x0 int, y0 int, x1 int, y1 int) {

	// Draw rectangle
	draw.Draw(cv, image.Rect(x0, y0, x1, y1), img, image.Point{}, draw.Src)

	// Blend background
	xgraphics.BlendBgColor(cv, color)
}

func drawText(cv *xgraphics.Image, txt string, color xgraphics.BGRA, x int, y int, size int) {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Error(err)
		return
	}

	// Obtain maximum font size
	w, _ := xgraphics.Extents(font, float64(size), txt)
	if w > 2*(x-fontMargin-rectMargin) {
		drawText(cv, txt, color, x, y, size-1)
		return
	}

	// Draw text onto canvas
	cv.Text(x-w/2, y-size, color, float64(size), font, txt)
}

func showGraphics(img *xgraphics.Image, ws *desktop.Workspace, duration time.Duration) *xwindow.Window {
	win, err := xwindow.Generate(img.X)
	if err != nil {
		log.Error(err)
		return nil
	}

	// Calculate window dimensions
	dim := xrect.New(store.DesktopDimensions(ws.Location.ScreenNum))
	w, h := img.Rect.Dx(), img.Rect.Dy()
	x, y := dim.X()+dim.Width()/2-w/2, dim.Y()+dim.Height()/2-h/2

	// Create the graphics window
	win.Create(img.X.RootWin(), x, y, w, h, 0)

	// Set class and name
	icccm.WmClassSet(win.X, win.Id, &icccm.WmClass{
		Instance: common.Build.Name,
		Class:    common.Build.Name,
	})
	icccm.WmNameSet(win.X, win.Id, common.Build.Name)

	// Set states for modal like behavior
	icccm.WmStateSet(win.X, win.Id, &icccm.WmState{
		State: icccm.StateNormal,
	})
	ewmh.WmStateSet(win.X, win.Id, []string{
		"_NET_WM_STATE_SKIP_TASKBAR",
		"_NET_WM_STATE_SKIP_PAGER",
		"_NET_WM_STATE_ABOVE",
		"_NET_WM_STATE_MODAL",
	})

	// Set hints for size and decorations
	icccm.WmNormalHintsSet(img.X, win.Id, &icccm.NormalHints{
		Flags:     icccm.SizeHintPPosition | icccm.SizeHintPMinSize | icccm.SizeHintPMaxSize,
		X:         x,
		Y:         y,
		MinWidth:  uint(w),
		MinHeight: uint(h),
		MaxWidth:  uint(w),
		MaxHeight: uint(h),
	})
	motif.WmHintsSet(img.X, win.Id, &motif.Hints{
		Flags:      motif.HintFunctions | motif.HintDecorations,
		Function:   motif.FunctionNone,
		Decoration: motif.DecorationNone,
	})

	// Ensure the window closes gracefully
	win.WMGracefulClose(func(w *xwindow.Window) {
		xevent.Detach(w.X, w.Id)
		xevent.Quit(w.X)
		w.Destroy()
	})

	// Paint the image and map the window
	img.XSurfaceSet(win.Id)
	img.XPaint(win.Id)
	img.XDraw()
	win.Map()

	// Close previous opened window
	if v, ok := gui[ws.Location.ScreenNum]; ok {
		v.Destroy()
	}
	gui[ws.Location.ScreenNum] = win

	// Close window after given duration
	if duration > 0 {
		time.AfterFunc(duration*time.Millisecond, win.Destroy)
	}

	return win
}

func scale(x, y, w, h int) (sx, sy, sw, sh int) {
	s := 10

	// Rescale dimensions by factor s
	sx, sy, sw, sh = x/s, y/s, w/s, h/s

	return
}

func bgra(name string) xgraphics.BGRA {
	rgba := common.Config.Colors[name]

	// Validate length
	if len(rgba) != 4 {
		log.Warn("Error obtaining color for ", name)
		return xgraphics.BGRA{}
	}

	return xgraphics.BGRA{
		R: uint8(rgba[0]),
		G: uint8(rgba[1]),
		B: uint8(rgba[2]),
		A: uint8(rgba[3]),
	}
}
