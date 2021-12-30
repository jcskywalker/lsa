package opencypher

type ResultSet struct {
	Rows [][]Value
}

func (r *ResultSet) Union(src ResultSet, all bool) error {
}
