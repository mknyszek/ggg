package ggg

import (
	"image/color"

	"github.com/golang/freetype/truetype"
)

type Theme struct {
	Name                  string
	ForegroundColor       color.Color
	GridlineColor         color.Color
	ChartBackgroundColor  color.Color
	BorderBackgroundColor color.Color
	SeriesPalette         func(uint64) color.Color
	TitleFont             *truetype.Font
	AxisFont              *truetype.Font
	AnnotationFont        *truetype.Font
}

var themes = make(map[string]*Theme)

func RegisterTheme(th *Theme) bool {
	if _, ok := themes[th.Name]; ok {
		return false
	}
	themes[th.Name] = th
	return true
}
