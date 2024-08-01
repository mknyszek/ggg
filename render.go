package ggg

import (
	"fmt"
	"image"
	"math"
	"slices"
	"strconv"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

func (p *Plot) Render(theme string, width, height int) (image.Image, error) {
	th, ok := themes[theme]
	if !ok {
		return nil, fmt.Errorf("unknown theme %s", theme)
	}
	c := gg.NewContext(width, height)

	// Basic dimensional definitions.
	w, h := float64(width), float64(height)
	var (
		padTop   = w / 10.0
		padLeft  = h / 8.0
		padRight = w / 20.0
		padBot   = h / 8.0
	)
	thickness := math.Round(math.Sqrt(w * h / (1080 * 720)))

	titleFont := truetype.NewFace(th.TitleFont, &truetype.Options{Size: math.Round(h / 15), SubPixelsX: 32, SubPixelsY: 8})
	axisFont := truetype.NewFace(th.AxisFont, &truetype.Options{Size: math.Round(h / 30), SubPixelsX: 32, SubPixelsY: 8})
	annotationFont := truetype.NewFace(th.AnnotationFont, &truetype.Options{Size: math.Round(h / 50), SubPixelsX: 32, SubPixelsY: 8})

	// Background color.
	c.DrawRectangle(0, 0, w, h)
	c.SetColor(th.BorderBackgroundColor)
	c.Fill()
	c.DrawRectangle(padLeft, padTop, w-padLeft-padRight, h-padTop-padBot)
	c.SetColor(th.ChartBackgroundColor)
	c.Fill()

	// Axis titles.
	c.SetFontFace(axisFont)
	c.SetColor(th.ForegroundColor)
	c.DrawStringWrapped(p.opts.x.title, (w-padLeft-padRight)/2+padLeft, h-padBot/2, 0.5, 0.5, w-padLeft-padRight, 8, gg.AlignCenter)
	c.Push()
	c.Translate(padLeft/4, (h-padTop-padBot)/2+padTop)
	c.Rotate(-math.Pi / 2)
	c.DrawStringWrapped(p.opts.y.title, 0, 0, 0.5, 0.5, w-padLeft-padRight, 8, gg.AlignCenter)
	c.Pop()

	// Plot title.
	c.SetFontFace(titleFont)
	c.SetColor(th.ForegroundColor)
	c.DrawStringWrapped(p.opts.title, padLeft/2, padTop/4, 0, 0, w-padRight, 8, gg.AlignLeft)

	// If there are no layers, there's nothing else to draw.
	if len(p.layers) == 0 {
		return c.Image(), nil
	}

	// Determine x/y ranges.
	if p.opts.x.userLimits {
		p.opts.x.min, p.opts.x.max = p.opts.x.userMin, p.opts.x.userMax
	} else {
		p.opts.x.min = math.Inf(1)
		p.opts.x.max = math.Inf(-1)
		for _, l := range p.layers {
			lo, hi := l.xRange()
			p.opts.x.min = min(p.opts.x.min, lo)
			p.opts.x.max = max(p.opts.x.max, hi)
		}
	}
	if p.opts.y.userLimits {
		p.opts.y.min, p.opts.y.max = p.opts.y.userMin, p.opts.y.userMax
	} else {
		p.opts.y.min = math.Inf(1)
		p.opts.y.max = math.Inf(-1)
		for _, l := range p.layers {
			lo, hi := l.yRange()
			p.opts.y.min = min(p.opts.y.min, lo)
			p.opts.y.max = max(p.opts.y.max, hi)
		}
	}

	// Set the scaling functions for x/y.
	var xScale, yScale scaleFunc
	if p.opts.x.logBase != 0 {
		if p.opts.x.min <= 0 || p.opts.x.max <= 0 {
			return nil, fmt.Errorf("specified log scale, but domain of X values is zero or negative: [%f, %f]", p.opts.x.min, p.opts.x.max)
		}
		xScale = scaleLog(p.opts.x.logBase, p.opts.x.min, p.opts.x.max, padLeft, w-padRight)
	} else {
		xScale = scaleLinear(p.opts.x.min, p.opts.x.max, padLeft, w-padRight)
	}
	if p.opts.y.logBase != 0 {
		if p.opts.y.min <= 0 || p.opts.y.max <= 0 {
			return nil, fmt.Errorf("specified log scale, but domain of Y values is zero or negative: [%f, %f]", p.opts.y.min, p.opts.y.max)
		}
		yScale = scaleLog(p.opts.y.logBase, p.opts.y.max, p.opts.y.min, padTop, h-padBot)
	} else {
		yScale = scaleLinear(p.opts.y.max, p.opts.y.min, padTop, h-padBot)
	}

	var xTicks, yTicks []float64
	if len(p.opts.x.customTicks) != 0 {
		xTicks = p.opts.x.customTicks
	} else if p.opts.x.logBase != 0 {
		xTicks = logTicks(p.opts.x.logBase, p.opts.x.min, p.opts.x.max)
	} else {
		xTicks = linearTicks(p.opts.x.min, p.opts.x.max, 5)
	}
	if len(p.opts.y.customTicks) != 0 {
		yTicks = p.opts.y.customTicks
	} else if p.opts.y.logBase != 0 {
		yTicks = logTicks(p.opts.y.logBase, p.opts.y.min, p.opts.y.max)
	} else {
		yTicks = linearTicks(p.opts.y.min, p.opts.y.max, 5)
	}

	// Draw gridlines.
	c.SetColor(th.GridlineColor)
	for _, x := range xTicks {
		dx := xScale(x)
		c.DrawLine(dx, h-padBot, dx, padTop)
	}
	for _, y := range yTicks {
		dy := yScale(y)
		c.DrawLine(padLeft, dy, w-padRight, dy)
	}
	c.Stroke()

	// Basic axes.
	c.SetLineCap(gg.LineCapSquare)
	c.SetLineJoin(gg.LineJoinBevel)
	c.SetLineWidth(2 * thickness)
	c.SetColor(th.ForegroundColor)
	c.DrawLine(padLeft, h-padBot, w-padRight, h-padBot)
	c.DrawLine(padLeft, h-padBot, padLeft, padTop)
	c.Stroke()

	// Draw ticks.
	c.SetColor(th.ForegroundColor)
	c.SetFontFace(annotationFont)
	for _, x := range xTicks {
		dx := xScale(x)
		c.DrawLine(dx, h-padBot, dx, h-padBot+padBot/10)
		c.DrawStringWrapped(strconv.FormatFloat(x, 'g', 3, 64), dx, h-padBot+padBot/5, 0.5, 0.5, (w-padLeft-padRight)/float64(len(xTicks)), 8, gg.AlignCenter)
	}
	for _, y := range yTicks {
		dy := yScale(y)
		c.DrawLine(padLeft, dy, padLeft-padLeft/10, dy)
		c.DrawStringWrapped(strconv.FormatFloat(y, 'g', 3, 64), padLeft-padLeft/5, dy, 1, 0.5, padLeft, 8, gg.AlignRight)
	}
	c.Stroke()

	// Draw layers.
	for _, l := range p.layers {
		im, err := l.render(th, width, height, xScale, yScale)
		if err != nil {
			return nil, err
		}
		c.DrawImage(im, 0, 0)
	}

	return c.Image(), nil
}

type scaleFunc func(float64) float64

func scaleLinear(x0, x1, t0, t1 float64) scaleFunc {
	m := (t1 - t0) / (x1 - x0)
	return func(a float64) float64 {
		return (a-x0)*m + t0
	}
}

func scaleLog(base int, x0, x1, t0, t1 float64) scaleFunc {
	log := logFunc(base)
	m := (t1 - t0) / (log(x1) - log(x0))
	return func(a float64) float64 {
		return (log(a)-log(x0))*m + t0
	}
}

func logFunc(base int) func(float64) float64 {
	switch base {
	case 2:
		return math.Log2
	case 10:
		return math.Log10
	}
	return func(x float64) float64 {
		return math.Log2(x) / math.Log2(float64(base))
	}
}

var (
	e10 = math.Sqrt(50)
	e5  = math.Sqrt(10)
	e2  = math.Sqrt(2)
)

func linearTicks(start, stop float64, count int) []float64 {
	if count <= 0 {
		count = 5
	}
	if start == stop {
		return []float64{start}
	}
	var i1, i2 int
	var inc float64
	if stop < start {
		i1, i2, inc = linearTickSpec(stop, start, count)
	} else {
		i1, i2, inc = linearTickSpec(start, stop, count)
	}
	if i2 < i1 {
		return nil
	}
	n := i2 - i1 + 1
	ticks := make([]float64, 0)
	if stop < start {
		if inc < 0 {
			for i := range n {
				ticks = append(ticks, float64(i2-i)/-inc)
			}
		} else {
			for i := range n {
				ticks = append(ticks, float64(i2-i)*inc)
			}
		}
	} else {
		if inc < 0 {
			for i := range n {
				ticks = append(ticks, float64(i1+i)/-inc)
			}
		} else {
			for i := range n {
				ticks = append(ticks, float64(i1+i)*inc)
			}
		}
	}
	return ticks
}

func linearTickSpec(start, stop float64, count int) (i1, i2 int, inc float64) {
	if count < 2 {
		count = 2
	}
	step := (stop - start) / float64(count)
	power := int(math.Floor(math.Log10(step)))
	err := step / math.Pow10(power)
	var factor float64
	switch {
	case err >= e10:
		factor = 10
	case err >= e5:
		factor = 5
	case err >= e2:
		factor = 2
	default:
		factor = 1
	}
	if power < 0 {
		inc = math.Pow10(-power) / factor
		i1 = int(math.Round(start * inc))
		i2 = int(math.Round(stop * inc))
		if float64(i1)/inc < start {
			i1++
		}
		if float64(i2)/inc > stop {
			i2--
		}
		inc = -inc
	} else {
		inc = math.Pow10(power) * factor
		i1 = int(math.Round(start / inc))
		i2 = int(math.Round(stop / inc))
		if float64(i1)*inc < start {
			i1++
		}
		if float64(i2)*inc > stop {
			i2--
		}
	}
	return
}

func logTicks(base int, start, stop float64) []float64 {
	log := logFunc(base)
	reverse := stop < start
	if reverse {
		start, stop = stop, start
	}
	lo := int(math.Floor(log(start)))
	hi := int(math.Ceil(log(stop)))
	majorTicks := hi - lo
	var ticksPerMajor int
	if base < 10 {
		ticksPerMajor = 1
	} else {
		ticksPerMajor = 4
	}
	ticks := make([]float64, 0, majorTicks*ticksPerMajor)
loop:
	for i := lo; i <= hi; i++ {
		major := math.Pow(float64(base), float64(i))
		for j := major / float64(ticksPerMajor); j <= major; j += major / float64(ticksPerMajor) {
			if j < start {
				continue
			}
			if j > stop {
				break loop
			}
			ticks = append(ticks, j)
		}
	}
	if reverse {
		slices.Reverse(ticks)
	}
	return ticks
}
