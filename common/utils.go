package common

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"crypto/sha1"
	"encoding/hex"

	"github.com/jezek/xgbutil/xrect"
)

type Pointer struct {
	X      int16  // Pointer X position relative to root
	Y      int16  // Pointer Y position relative to root
	Button uint16 // Pointer button states of device
}

func Hash(text string) string {
	hash := sha1.New()
	hash.Write([]byte(text))
	return hex.EncodeToString(hash.Sum(nil))
}

func Truncate(s string, max int) string {
	if max > len(s) {
		return s
	}
	return s[:max]
}

func IsType(a interface{}, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}

func IsZero(items []uint) bool {
	mask := uint(0)
	for _, s := range items {
		mask |= s
	}
	return mask == 0
}

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

func VersionToInt(version string) int {

	// Remove non-numeric characters
	reg := regexp.MustCompile("[^0-9]+")
	numeric := reg.ReplaceAllString(strings.Split(version, "-")[0], "")

	// Convert version string to integer
	integer, err := strconv.Atoi(numeric)
	if err != nil {
		return -1
	}

	return integer
}
