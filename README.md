# ggg

Data manipulation and plotting package for Go.

```go
package main

import (
	"image/png"
	"log"
	"math"
	"os"

	. "github.com/mknyszek/ggg"
	_ "github.com/mknyszek/ggg/themes/dark"
)

func main() {
    // Create a simple dataset.
	d := Empty()
	cx, _ := Extend[int](d, "x")
	cy, _ := Extend[float64](d, "y")
	i := 0
	for row := range Append(d, 100) {
		SetField(row, cx, i)
		SetField(row, cy, math.Cos(float64(i)/100.0))
		i++
	}

    // Plot it.
	p := NewPlot().Layer(
        &Layer1D[int, float64]{
            Data: d,
            X:    cx,
            Y:    cy,
            Geom: Line(PaletteColor(0), Constant(2.0)),
        },
    ).Layer(
        &Layer1D[int, float64]{
            Data: d,
            X:    cx,
            Y:    cy,
            Geom: Points(PaletteColor(0), Constant(2.0)),
        },
    ).Presentation(
        Title("My Chart"),
        XAxis("boxes", LogScale(10)),
        YAxis("tons of fish"),
    )

    // Render and write out the plot.
    im, err := p.Render("dark", 2160, 1440)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create("./out.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, im); err != nil {
		log.Fatal(err)
	}
}
```

## Goals

- Flexible data model.
- Type safety.
- Grammer-of-graphics-style visualization.

## Status

Experimental work-in-progress. Don't expect backwards-compatibility.

## Possible future features

Data model:
- Parsing CSV/TSV.
- Deleting rows and columns.
- Joins across datasets.
- Row-to-column transforms.

Visualization:
- 2D layers (areas, etc.).
- Bar geom.
- Legends.
- Point shapes.
- Faceting.
