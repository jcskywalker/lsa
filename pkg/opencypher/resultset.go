package opencypher

import (
	"errors"
	"reflect"
)

var ErrRowsHaveDifferentSizes = errors.New("Rows have different sizes")
var ErrIncompatibleCells = errors.New("Incompatible result set cells")

// ResultSet is a table of values
type ResultSet struct {
	Rows [][]Value
}

func isCompatibleValue(v1, v2 Value) bool {
	if v1.Value == nil {
		if v2.Value == nil {
			return true
		}
		return false
	}
	if v2.Value == nil {
		return false
	}
	return reflect.TypeOf(v1.Value) == reflect.TypeOf(v2.Value)
}

// Check if all row cells are compatible
func isCompatibleRow(row1, row2 []Value) error {
	if len(row1) != len(row2) {
		return ErrRowsHaveDifferentSizes
	}
	for i := range row1 {
		if !isCompatibleValue(row1[i], row2[i]) {
			return ErrIncompatibleCells
		}
	}
	return nil
}

func (r *ResultSet) find(row []Value) int {
	for index, r := range r.Rows {
		if len(r) != len(row) {
			break
		}
		found := true
		for i := range r {
			if !r[i].IsSame(row[i]) {
				found = false
				break
			}
		}
		if found {
			return index
		}
	}
	return -1
}

// Append the row to the resultset. The row must be compatible (i.e. same types in every cell)
func (r *ResultSet) Append(row []Value) error {
	if len(r.Rows) != 0 {
		if err := isCompatibleRow(r.Rows[0], row); err != nil {
			return err
		}
	}
	r.Rows = append(r.Rows, row)
	return nil
}

// Union adds the src resultset to this. If all is set, it adds all rows, otherwise, it adds unique rows
func (r *ResultSet) Union(src ResultSet, all bool) error {
	for _, sourceRow := range src.Rows {
		appnd := all
		if !appnd && r.find(sourceRow) != -1 {
			appnd = true
		}
		if appnd {
			if err := r.Append(sourceRow); err != nil {
				return err
			}
		}
	}
	return nil
}
