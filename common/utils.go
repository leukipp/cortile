package common

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

func IsInList(item string, items []string) bool {
	for i := 0; i < len(items); i++ {
		if items[i] == item {
			return true
		}
	}
	return false
}

func IsInsideRect(p *Pointer, r xrect.Rect) bool {
	x, y, w, h := r.Pieces()

	// Check if x and y are inside rectangle
	xInRect := int(p.X) >= x && int(p.X) <= (x+w)
	yInRect := int(p.Y) >= y && int(p.Y) <= (y+h)

	return xInRect && yInRect
}
