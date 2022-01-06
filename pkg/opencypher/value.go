package opencypher

import (
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

// Value represents a computer value. Possible data types it can contain are:
//
//   primitives:
//    int
//    float64
//    bool
//    string
//    neo4j.Duration
//    neo4j.Date
//    neo4j.LocalDateTime
//    neo4j.LocalTime
//
//  composites:
//    []Value
//    map[string]Value
//    Labels
//    OCNode
//    OCEdge
//    OCPath
type Value struct {
	Value    interface{}
	Constant bool
}

// OCNode represents a node object as viewed by Opencypher
type OCNode interface {
	// GetProperty returns the value of a property
	GetProperty(string) (Value, bool)
	// Returns the labels of the node
	GetLabels() Labels
	// SameNode returns if the underlying nodes are the same node
	SameNode(OCNode) bool
}

// OCEdge represents an edge object
type OCEdge interface {
	GetTypes() []string
	GetProperty(string) (Value, bool)
	SameEdge(OCEdge) bool
	GetFrom() OCNode
	GetTo() OCNode
}

// An OCPath is a list of edges
type OCPath []OCEdge

// IsPrimitive returns true if the value is int, float64, bool,
// string, duration, date, datetime, localDateTime, or localTime
func (v Value) IsPrimitive() bool {
	switch v.Value.(type) {
	case int, float64, bool, string, neo4j.Duration, neo4j.Date, neo4j.LocalDateTime, neo4j.LocalTime:
		return true
	}
	return false
}

// IsSame compares two values and decides if the two are the same
func (v Value) IsSame(v2 Value) bool {
	if v.IsPrimitive() {
		if v2.IsPrimitive() {
			eq, err := comparePrimitiveValues(v.Value, v2.Value)
			return err != nil && eq == 0
		}
		return false
	}

	switch val1 := v.Value.(type) {
	case []Value:
		val2, ok := v2.Value.([]Value)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for i := range val1 {
			if !val1[i].IsSame(val2[i]) {
				return false
			}
		}
		return true

	case map[string]Value:
		val2, ok := v2.Value.(map[string]Value)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for k, v := range val1 {
			v2, ok := val2[k]
			if !ok {
				return false
			}
			if !v.IsSame(v2) {
				return false
			}
		}
		return true

	case Labels:
		val2, ok := v2.Value.(Labels)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for k := range val1 {
			if _, exists := val2[k]; !exists {
				return false
			}
		}
		return true

	case OCNode:
		val2, ok := v2.Value.(OCNode)
		if !ok {
			return false
		}
		return val1.SameNode(val2)

	case OCEdge:
		val2, ok := v2.Value.(OCEdge)
		if !ok {
			return false
		}
		return val1.SameEdge(val2)

	case OCPath:
		val2, ok := v2.Value.(OCPath)
		if !ok {
			return false
		}
		if len(val1) != len(val2) {
			return false
		}
		for _, v1 := range val1 {
			found := false
			for _, v2 := range val2 {
				if v1.SameEdge(v2) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

func (v Value) Evaluate(ctx *EvalContext) (Value, error) { return v, nil }

// Labels is simply a set of string labels
type Labels map[string]struct{}
