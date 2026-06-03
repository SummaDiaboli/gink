package gink

import "strings"

// TreeNode is a single node in a [NewTree]. Children is nil or empty for leaf
// nodes. Expanded controls whether the node's children are visible; the tree
// mutates this field directly when the user expands or collapses a node.
type TreeNode struct {
	Label    string
	Children []TreeNode
	Expanded bool
}

// flatNode is an entry in the flattened visible-node list.
type flatNode struct {
	node  *TreeNode
	depth int
}

// flattenTree returns a depth-first list of currently visible nodes.
func flattenTree(nodes []TreeNode, depth int) []flatNode {
	var result []flatNode
	for i := range nodes {
		n := &nodes[i]
		result = append(result, flatNode{node: n, depth: depth})
		if n.Expanded && len(n.Children) > 0 {
			result = append(result, flattenTree(n.Children, depth+1)...)
		}
	}
	return result
}

// NewTree returns a focusable hierarchical tree component.
//
// nodes is the root-level slice. onSelect is called with a pointer to the
// TreeNode when the user presses Enter on a node. The tree mutates
// [TreeNode.Expanded] in place when the user expands or collapses a node,
// which causes the visible list to recompute on the next render.
//
// Navigation: Up/Down move between visible nodes; Right expands a collapsed
// parent node; Left collapses an expanded parent node. Enter fires onSelect.
//
// Parents render with a ▶ (collapsed) or ▼ (expanded) prefix. Leaves render
// with no prefix. Each depth level is indented by two spaces.
//
//	nodes := []gink.TreeNode{
//	    {Label: "src", Expanded: true, Children: []gink.TreeNode{
//	        {Label: "main.go"},
//	        {Label: "util.go"},
//	    }},
//	    {Label: "go.mod"},
//	}
//	gink.C(gink.NewTree(nodes, func(n *gink.TreeNode) { open(n.Label) }))
func NewTree(nodes []TreeNode, onSelect func(*TreeNode), styles ...Style) func() Element {
	explicitStyle, hasExplicitStyle := optionalStyle(styles)
	return func() Element {
		focusStyle := resolveStyle(explicitStyle, hasExplicitStyle, UseTheme().Focused)

		cursor, setCursor := UseState(0)
		isFocused := UseFocus()

		visible := flattenTree(nodes, 0)

		// Clamp cursor in case the tree shrinks (e.g. after a collapse).
		if cursor >= len(visible) {
			cursor = len(visible) - 1
			setCursor(cursor)
		}
		if cursor < 0 {
			cursor = 0
		}

		UseInput(func(ev KeyEvent) {
			if !isFocused || len(visible) == 0 {
				return
			}
			cur := visible[cursor]
			switch ev.Key {
			case KeyUp:
				if cursor > 0 {
					setCursor(cursor - 1)
				}
			case KeyDown:
				if cursor < len(visible)-1 {
					setCursor(cursor + 1)
				}
			case KeyRight:
				if len(cur.node.Children) > 0 && !cur.node.Expanded {
					cur.node.Expanded = true
					setCursor(cursor) // trigger re-render
				}
			case KeyLeft:
				if cur.node.Expanded {
					cur.node.Expanded = false
					setCursor(cursor)
				}
			case KeyEnter:
				onSelect(cur.node)
			}
		})

		UseClick(func(_, localY int) {
			if localY >= 0 && localY < len(visible) {
				setCursor(localY)
				onSelect(visible[localY].node)
			}
		})

		const (
			treeExpanded  = "▼ "
			treeCollapsed = "▶ "
			treeLeaf      = "  "
		)
		rows := make([]Element, len(visible))
		for i, fn := range visible {
			indent := strings.Repeat("  ", fn.depth)
			var prefix string
			if len(fn.node.Children) > 0 {
				if fn.node.Expanded {
					prefix = treeExpanded
				} else {
					prefix = treeCollapsed
				}
			} else {
				prefix = treeLeaf
			}
			style := NewStyle()
			if i == cursor && isFocused {
				style = focusStyle
			}
			rows[i] = Text(indent+prefix+fn.node.Label, style)
		}
		return Box(rows...)
	}
}
