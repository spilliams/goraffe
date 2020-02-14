package grapher

type GraphFormat int

const UnknownGraphFormat = "unknown graph format"

func (gf GraphFormat) String() string {
	switch gf {
	case Graphviz:
		return "graphviz"
	}
	return UnknownGraphFormat
}

const (
	Graphviz GraphFormat = iota
)
