package gink

import (
	"strings"
	"testing"
)

// treeFixture returns a simple 3-level tree:
//
//	▼ Root
//	  ▶ Child A   (collapsed, has children)
//	    Leaf A1
//	    Leaf A2
//	  Leaf B      (no children)
func treeFixture() []TreeNode {
	return []TreeNode{
		{
			Label:    "Root",
			Expanded: true,
			Children: []TreeNode{
				{
					Label: "Child A",
					Children: []TreeNode{
						{Label: "Leaf A1"},
						{Label: "Leaf A2"},
					},
				},
				{Label: "Leaf B"},
			},
		},
	}
}

func treeHarness(t *testing.T, nodes []TreeNode) (*Harness, *TreeNode) {
	t.Helper()
	var lastSel *TreeNode
	h := NewHarness(t, func() Element {
		return C(NewTree(nodes, func(n *TreeNode) { lastSel = n }))
	})
	return h, lastSel
}

// ── Rendering ────────────────────────────────────────────────────────────────

func TestNewTree_rendersRootNode(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	if !h.Contains("Root") {
		t.Errorf("root node not rendered; lines: %v", h.Lines())
	}
}

func TestNewTree_expandedNodeShowsChildren(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Root is expanded; Child A and Leaf B should be visible.
	if !h.Contains("Child A") {
		t.Error("expanded root should show Child A")
	}
	if !h.Contains("Leaf B") {
		t.Error("expanded root should show Leaf B")
	}
}

func TestNewTree_collapsedNodeHidesChildren(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Child A is collapsed; its children should not be visible.
	if h.Contains("Leaf A1") {
		t.Error("collapsed Child A should hide Leaf A1")
	}
}

func TestNewTree_expandIndicatorOnParent(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	for _, line := range h.Lines() {
		if strings.Contains(line, "Child A") {
			if !strings.Contains(line, "▶") {
				t.Errorf("collapsed parent should show ▶; line: %q", line)
			}
			return
		}
	}
	t.Fatal("Child A not found")
}

func TestNewTree_collapseIndicatorOnExpandedParent(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	for _, line := range h.Lines() {
		if strings.Contains(line, "Root") {
			if !strings.Contains(line, "▼") {
				t.Errorf("expanded parent should show ▼; line: %q", line)
			}
			return
		}
	}
	t.Fatal("Root not found")
}

func TestNewTree_leafNodeNoIndicator(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	for _, line := range h.Lines() {
		if strings.Contains(line, "Leaf B") {
			if strings.Contains(line, "▶") || strings.Contains(line, "▼") {
				t.Errorf("leaf node should have no expand indicator; line: %q", line)
			}
			return
		}
	}
	t.Fatal("Leaf B not found")
}

// ── Navigation ───────────────────────────────────────────────────────────────

func TestNewTree_downMovesToNextVisibleNode(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	h.SendKey(KeyDown)

	// Cursor moved to Child A — it should be highlighted.
	for i, line := range h.Lines() {
		if strings.Contains(line, "Child A") {
			if h.CellStyle(0, i) == (Style{}).toTcell() {
				t.Error("Child A should be highlighted after Down")
			}
			return
		}
	}
	t.Fatal("Child A not found")
}

func TestNewTree_upNoOpAtFirstNode(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Root is cursor at start; Up should be no-op (Root stays highlighted).
	h.SendKey(KeyUp)

	for i, line := range h.Lines() {
		if strings.Contains(line, "Root") {
			if h.CellStyle(0, i) == (Style{}).toTcell() {
				t.Error("Root should still be highlighted after Up at first node")
			}
			return
		}
	}
	t.Fatal("Root not found")
}

// ── Expand / Collapse ────────────────────────────────────────────────────────

func TestNewTree_rightExpandsCollapsedNode(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Move down to Child A (collapsed) and expand with Right.
	h.SendKey(KeyDown)
	h.SendKey(KeyRight)

	if !h.Contains("Leaf A1") {
		t.Error("Right on collapsed node should expand it and show children")
	}
}

func TestNewTree_leftCollapsesExpandedNode(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Root is expanded — Left should collapse it.
	h.SendKey(KeyLeft)

	if h.Contains("Child A") {
		t.Error("Left on expanded node should collapse it and hide children")
	}
}

func TestNewTree_rightNoOpOnLeaf(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	// Navigate to Leaf B (Down twice from Root).
	h.SendKey(KeyDown) // Child A
	h.SendKey(KeyDown) // Leaf B
	h.SendKey(KeyRight)

	// Nothing should change — no crash, no extra nodes.
	if h.Contains("Leaf A1") {
		t.Error("Right on leaf should be no-op; did not expect Leaf A1 to appear")
	}
}

// ── Selection ────────────────────────────────────────────────────────────────

func TestNewTree_enterFiresOnSelect(t *testing.T) {
	var lastSel *TreeNode
	h := NewHarness(t, func() Element {
		nodes := treeFixture()
		return C(NewTree(nodes, func(n *TreeNode) { lastSel = n }))
	})
	defer h.Close()

	h.SendKey(KeyEnter)

	if lastSel == nil || lastSel.Label != "Root" {
		t.Errorf("Enter should select current node; got %v", lastSel)
	}
}

func TestNewTree_indentsChildrenByDepth(t *testing.T) {
	h, _ := treeHarness(t, treeFixture())
	defer h.Close()

	var rootCol, childCol int
	for _, line := range h.Lines() {
		runes := []rune(line)
		if strings.Contains(line, "Root") {
			for i, r := range runes {
				if r != ' ' {
					rootCol = i
					break
				}
			}
		}
		if strings.Contains(line, "Child A") {
			for i, r := range runes {
				if r != ' ' {
					childCol = i
					break
				}
			}
		}
	}
	if childCol <= rootCol {
		t.Errorf("Child A (col %d) should be indented more than Root (col %d)", childCol, rootCol)
	}
}
