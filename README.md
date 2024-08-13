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

var (
	colX = NewColumn[int]("x")
	colY = NewColumn[float64]("y")
	colS = NewColumn[string]("name")
)

func main() {
	// Create a simple dataset.
	d := Empty()
	d.AddColumn(colX)
	d.AddColumn(colY)
	d.AddColumn(colS)

	i := 1
	for row := range d.Grow(100) {
		colS.Set(d, row, "mackeral")
		colX.Set(d, row, i)
		colY.Set(d, row, math.Cos(float64(i)/100.0))
		i++
	}
	i = 1
	for row := range d.Grow(100) {
		colS.Set(d, row, "herring")
		colX.Set(d, row, i)
		colY.Set(d, row, math.Sin(float64(i)/100.0))
		i++
	}

	// Plot it.
	p := LinePlot(d, colX, colY, colS).Presentation(
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
- Deleting rows.
- Joins across datasets.
- Pivoting.

Visualization:
- 2D layers (areas, etc.).
- Bar geom.
- Legends.
- Point shapes.
- Faceting.
