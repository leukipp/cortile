package common

import (
	"strconv"
	"strings"

	"crypto/sha1"
	"encoding/hex"

	"github.com/jezek/xgbutil/xrect"
)

type Point struct {
	X int // Object point x position
	Y int // Object point y position
}

func CreatePoint(x int, y int) *Point {
	return &Point{
		X: x,
		Y: y,
	}
}

type Geometry struct {
	X      int // Object geometry x position
	Y      int // Object geometry y position
	Width  int // Object geometry width dimension
	Height int // Object geometry height dimension
}

func CreateGeometry(r xrect.Rect) *Geometry {
	return &Geometry{
		X:      r.X(),
		Y:      r.Y(),
		Width:  r.Width(),
		Height: r.Height(),
	}
}

func (g *Geometry) Center() Point {
	return *CreatePoint(g.X+g.Width/2, g.Y+g.Height/2)
}

func (g *Geometry) Rect() xrect.Rect {
	return xrect.New(g.X, g.Y, g.Width, g.Height)
}

func (g *Geometry) Pieces() (int, int, int, int) {
	return g.X, g.Y, g.Width, g.Height
}

type Map = map[string]interface{} // Generic map type
type List = []Map                 // Generic list type

func HashString(text string, max int) string {
	hash := sha1.New()
	hash.Write([]byte(text))
	str := hex.EncodeToString(hash.Sum(nil))
	return TruncateString(str, max)
}

func TruncateString(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:max]
}

func RemoveChars(s string, chars []string) string {
	for _, c := range chars {
		s = strings.Replace(s, c, "", -1)
	}
	return s
}

func AllZero(items []uint) bool {
	mask := uint(0)
	for _, item := range items {
		mask |= item
	}
	return mask == 0
}

func AllTrue(items []bool) bool {
	mask := true
	for _, item := range items {
		mask = mask && item
	}
	return mask
}

func IsInsideRect(p Point, g Geometry) bool {
	x, y, w, h := g.Pieces()
	xInRect := int(p.X) >= x && int(p.X) <= (x+w)
	yInRect := int(p.Y) >= y && int(p.Y) <= (y+h)
	return xInRect && yInRect
}

func IsInList(item string, items []string) bool {
	for i := 0; i < len(items); i++ {
		if items[i] == item {
			return true
		}
	}
	return false
}

func IsInMap(m Map, keys []string) bool {
	exists := true
	for _, key := range keys {
		_, exist := m[key]
		exists = exists && exist
	}
	return exists
}

func ReverseList[T any](items []T) []T {
	for i, j := 0, len(items)-1; i < j; {
		items[i], items[j] = items[j], items[i]
		i++
		j--
	}
	return items
}

func StringsToInts(items []string) []int {
	result := make([]int, len(items))
	for i, item := range items {
		integer, err := strconv.Atoi(item)
		if err != nil {
			integer = -1
		}
		result[i] = integer
	}
	return result
}
