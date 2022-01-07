package opencypher

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

// A NodeList is a view on a graph that selects a subset of the nodes
type NodeList interface {
	// ScanNodes scans all the nodes in the current graph based on the
	// given labels and properties, and returns a new graph containing
	// only those nodes
	ScanNodes(labels map[string]struct{}, properties map[string]Value) (NodeList, error)
}

type EdgeList interface {
	// Scan edges in the current edge list, and return a new edge list
	// If start or and is not null, start and end nodes are matched to those
	// labels specify the labels to search for.
	ScanEdges(start, end OCNode, labels map[string]struct{}, properties map[string]Value) (EdgeList, error)
}

// OCGraph is how this interpreter sees a graph
type OCGraph interface {
	NodeList
	EdgeList
}
