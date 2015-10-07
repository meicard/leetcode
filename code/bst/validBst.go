package validBst

type treeNode struct {
	left  *treeNode
	right *treeNode
	val   int
}

func validBst(root *treeNode) bool {
	if root == nil {
		return true
	}

	return subTreeLessThan(root.left, root.val) &&
		subTreeGreaterThan(root.right, root.val) &&
		validBst(root.left) &&
		validBst(root.right)
}

func subTreeLessThan(p *treeNode, int val) bool {
	if p == nil {
		return true
	}

	return p.val < val &&
		subTreeLessThan(p.left, val) &&
		subTreeLessThan(p.right, val)
}

func subTreeGreaterThan(p *treeNode, int val) bool {
	if p == nil {
		return true
	}

	return p.val > val &&
		subTreeGreaterThan(p.left, val) &&
		subTreeGreaterThan(p.right, val)
}
