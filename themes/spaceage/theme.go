package spaceage

import (
	_ "embed"
	"fmt"
	"image/color"

	"github.com/golang/freetype/truetype"
	"github.com/mknyszek/ggg"
	"github.com/mknyszek/ggg/third_party/spacegrotesk"
)

func init() {
	f, err := truetype.Parse(spacegrotesk.LightTTF)
	if err != nil {
		panic(err)
	}
	ok := ggg.RegisterTheme(&ggg.Theme{
		Name:           "spaceage",
		TitleFont:      f,
		AxisFont:       f,
		AnnotationFont: f,
		SeriesPalette: func(i uint64) color.Color {
			return palette[int(i)%len(palette)]
		},
		BorderBackgroundColor: &color.RGBA{235, 223, 211, 255},
		ChartBackgroundColor:  &color.RGBA{245, 236, 225, 255},
		ForegroundColor:       &color.RGBA{41, 41, 41, 255},
		GridlineColor:         &color.RGBA{41, 41, 41, 64},
	})
	if !ok {
		panic(fmt.Sprintf("theme 'spaceage' already exists"))
	}
}

var palette = []color.Color{
	&color.RGBA{0xd5, 0x3e, 0x4f, 0xff},
	&color.RGBA{0xe8, 0x5a, 0x48, 0xff},
	&color.RGBA{0xf6, 0x7a, 0x49, 0xff},
	&color.RGBA{0xfb, 0xa1, 0x5b, 0xff},
	&color.RGBA{0xfd, 0xc2, 0x72, 0xff},
	&color.RGBA{0xfe, 0xe0, 0x8b, 0xff},
	&color.RGBA{0xe6, 0xf5, 0x98, 0xff},
	&color.RGBA{0xba, 0xe3, 0xa1, 0xff},
	&color.RGBA{0x89, 0xd0, 0xa5, 0xff},
	&color.RGBA{0x59, 0xb4, 0xab, 0xff},
	&color.RGBA{0x32, 0x88, 0xbd, 0xff},
}
