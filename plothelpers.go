package ggg

// LinePlot is a helper to create a simple line plot where the values of the series column
// determine how to group the data.
func LinePlot[X, Y Scalar, S comparable](d *Dataset, x Column[X], y Column[Y], series Column[S]) *Plot {
	return NewPlot().Layer(
		&Layer[X, Y]{
			Data: d,
			X:    x,
			Y:    y,
			Geom: Line(NiceColors(series), Constant(2.0)),
		},
	)
}
