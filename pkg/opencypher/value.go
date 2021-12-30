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
//    Labels
type Value struct {
	Value    interface{}
	Constant bool
}

func (v Value) Evaluate(ctx *EvalContext) (Value, error) { return v, nil }

type Labels map[string]struct{}
