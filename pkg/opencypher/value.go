package opencypher

// Value represents a computer value. Possible data types it can contain are:
//
//    nil
//    int
//    float64
//    bool
//    string
//    []Value
//    map[string]Value
//    neo4j.Duration
//    neo4j.Date
//    neo4j.DateTime
//    neo4j.LocalDateTime
//    neo4j.LocalTime
//    GraphObject
type Value struct {
	Value    interface{}
	Constant bool
}

func (v Value) Evaluate(ctx *EvalContext) (Value, error) { return v, nil }

type GraphObjectType int

const (
	GraphObjectUnknown GraphObjectType = 0
	GraphObjectNode    GraphObjectType = 1
	GraphObjectEdge    GraphObjectType = 2
)

type GraphObject struct {
	Type       GraphObjectType
	Labels     []string
	Properties map[string]Value
}

func (g GraphObject) HasLabel(label string) bool {
	for _, x := range g.Labels {
		if label == x {
			return true
		}
	}
	return false
}
