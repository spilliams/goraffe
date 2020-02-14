package grapher

import (
	"fmt"
)

type Digraph interface{}

func New() *Grapher {
	return &Grapher{
		digraphs: make([]Digraph, 0),
	}
}

type Grapher struct {
	format   GraphFormat
	digraphs []Digraph
}

func (g *Grapher) SetFormat(gf GraphFormat) error {
	if gf.String() == UnknownGraphFormat {
		return fmt.Errorf("%s %v", gf, gf)
	}
	g.format = gf
	return nil
}

func (g *Grapher) AddDigraph(d Digraph) {
	g.digraphs = append(g.digraphs, d)
}

func (g *Grapher) Render() ([]byte, error) {
	// TODO
	return []byte{}, nil
}
