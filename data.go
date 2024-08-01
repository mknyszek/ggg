package ggg

import (
	"fmt"
	"iter"
	"reflect"
	"unique"
)

// Dataset represents some data. The data are structured as a series
// of rows with uniformly-typed columns.
type Dataset struct {
	columns []columnI
	rows    int
	colMap  map[unique.Handle[columnKey]]int
}

type columnData[T any] struct {
	name   string
	key    unique.Handle[columnKey]
	values []T
}

func (c *columnData[T]) id() string {
	return c.name
}

func (c *columnData[T]) grow(n int) {
	c.values = append(c.values, make([]T, n)...)
}

type columnI interface {
	id() string
	grow(n int)
}

// Empty returns an empty dataset.
func Empty() *Dataset {
	return &Dataset{colMap: make(map[unique.Handle[columnKey]]int)}
}

// Grow adds new rows to the dataset and returns an iterator producing
// those new rows. Returns an iterator over the new row indices.
func (d *Dataset) Grow(n int) iter.Seq[int] {
	s := d.rows
	d.rows += n
	for _, c := range d.columns {
		c.grow(n)
	}
	return func(yield func(int) bool) {
		for i := s; i < s+n; i++ {
			if !yield(i) {
				break
			}
		}
	}
}

// Column represents a column of uniformly-typed values in a particular Dataset.
type Column[T any] struct {
	key   unique.Handle[columnKey]
	cache *int
}

// NewColumn returns a new column with the provided name that may be used to access
// and mutate a dataset.
func NewColumn[T any](name string) Column[T] {
	return Column[T]{cache: new(int), key: unique.Make(columnKey{name: name, typ: reflect.TypeFor[T]()})}
}

type columnKey struct {
	name string
	typ  reflect.Type
}

// String returns a debug string for the column.
func (c Column[T]) String() string {
	v := c.key.Value()
	return fmt.Sprintf("%s (%s)", v.name, v.typ)
}

// All returns an iterator over all values in the column in the dataset.
func (c Column[T]) All(d *Dataset) iter.Seq[T] {
	var colData *columnData[T]
	if cd, ok := d.columns[*c.cache].(*columnData[T]); ok && cd.key == c.key {
		// Fast path: our cache has the right index.
		colData = cd
	} else {
		ci, ok := d.colMap[c.key]
		if !ok {
			panic(fmt.Sprintf("column %s not in dataset", c))
		}
		*c.cache = ci
		colData = d.columns[ci].(*columnData[T])
	}
	return func(yield func(T) bool) {
		for _, value := range colData.values {
			if !yield(value) {
				break
			}
		}
	}
}

// Get retrieves a value in the dataset at a particular row for this column.
func (c Column[T]) Get(d *Dataset, row int) T {
	// Fast path: our cache has the right index.
	if cd, ok := d.columns[*c.cache].(*columnData[T]); ok && cd.key == c.key {
		return cd.values[row]
	}
	return c.getSlow(d, row)
}

//go:noinline
func (c Column[T]) getSlow(d *Dataset, row int) T {
	ci, ok := d.colMap[c.key]
	if !ok {
		panic(fmt.Sprintf("column %s not in dataset", c))
	}
	*c.cache = ci
	return d.columns[ci].(*columnData[T]).values[row]
}

// Set sets a value in the dataset at a particular row for this column.
func (c Column[T]) Set(d *Dataset, row int, value T) {
	// Fast path: our cache has the right index.
	if cd, ok := d.columns[*c.cache].(*columnData[T]); ok && cd.key == c.key {
		cd.values[row] = value
	}
	c.setSlow(d, row, value)
}

//go:noinline
func (c Column[T]) setSlow(d *Dataset, row int, value T) {
	ci, ok := d.colMap[c.key]
	if !ok {
		panic(fmt.Sprintf("column %s not in dataset", c))
	}
	d.columns[ci].(*columnData[T]).values[row] = value
}

// Name returns the name of the column.
func (c Column[T]) Name() string {
	return c.key.Value().name
}

func (c Column[T]) colKey() unique.Handle[columnKey] {
	return c.key
}

func (c Column[T]) newData(rows int) columnI {
	return &columnData[T]{c.key.Value().name, c.key, make([]T, rows)}
}

// AnyColumn is a way to refer to Column[T] for all T.
type AnyColumn interface {
	Name() string

	colKey() unique.Handle[columnKey]
	newData(rows int) columnI
}

// AddColumn adds a new column to the dataset's structure. If the dataset already has
// rows, the column's data will be zero-initialized.
func (d *Dataset) AddColumn(c AnyColumn) bool {
	key := c.colKey()
	if _, ok := d.colMap[key]; ok {
		return false
	}
	d.columns = append(d.columns, c.newData(d.rows))
	d.colMap[key] = len(d.columns) - 1
	return true
}

// Columns returns the number of columns in the dataset.
func (d *Dataset) Columns() int {
	return len(d.columns)
}

// ColumnNames returns an iterator over the names of columns in the dataset.
func (d *Dataset) ColumnNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, c := range d.columns {
			if !yield(c.id()) {
				break
			}
		}
	}
}

// Rows returns the number of rows in the dataset.
func (d *Dataset) Rows() int {
	return d.rows
}
