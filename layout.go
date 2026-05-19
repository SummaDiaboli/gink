package gink

// Direction controls how a Box lays out its children.
type Direction int

const (
	DirectionColumn Direction = iota // vertical stacking (default)
	DirectionRow                     // horizontal stacking
)

// BoxProps holds layout metadata for a box element.
// Children are stored in Element.Children, not here.
type BoxProps struct {
	Direction Direction
	Gap       int // empty cells between children
}
