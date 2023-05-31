package common

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	log "github.com/sirupsen/logrus"
)

func IsInList(item string, items []string) bool {
	for i := 0; i < len(items); i++ {
		if items[i] == item {
			return true
		}
	}
	return false
}

func IsInsideRect(p *xproto.QueryPointerReply, r xrect.Rect) bool {
	x, y, w, h := r.Pieces()

	// Check if x and y are inside rectangle
	xInRect := int(p.RootX) >= x && int(p.RootX) <= (x+w)
	yInRect := int(p.RootY) >= y && int(p.RootY) <= (y+h)

	return xInRect && yInRect
}

func GetColor(name string) xgraphics.BGRA {
	rgba := Config.Colors[name]

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
