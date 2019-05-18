package dotprinter

import (
	"fmt"
	"strings"
)

const graphTypeDigraph = "digraph"

type Dotfile struct {
	graphType string
	name      string
	nodes     []node
	edges     []edge
}

type node struct {
	name       string
	attributes []string
}

func (n node) String() string {
	r := fmt.Sprintf("%s", n.name)
	if len(n.attributes) > 0 {
		r += fmt.Sprintf(" [%s]", strings.Join(n.attributes, ","))
	}
	return r
}

type edge struct {
	left       string
	operation  string
	right      string
	attributes []string
}

func (e edge) String() string {
	r := fmt.Sprintf("%s %s %s", e.left, e.operation, e.right)
	if len(e.attributes) > 0 {
		r += fmt.Sprintf(" [%s]", strings.Join(e.attributes, ","))
	}
	return r
}

func NewDigraph(name string) *Dotfile {
	d := Dotfile{
		graphType: graphTypeDigraph,
		name:      name,
		nodes:     make([]node, 0, 0),
		edges:     make([]edge, 0, 0),
	}

	return &d
}

func (d *Dotfile) String() string {
	r := fmt.Sprintf("%s %s {\n", d.graphType, d.name)

	for _, node := range d.nodes {
		r += fmt.Sprintf("\t%s;\n", node)
	}

	for _, edge := range d.edges {
		r += fmt.Sprintf("\t%s;\n", edge)
	}

	r += fmt.Sprintf("}\n")

	return r
}
