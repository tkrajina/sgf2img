package sgfutils

import (
	"github.com/rooklift/sgf"
)

type SubOpts struct{}

// If nodesAround < 0 => nodes *before* otherwise nodes *after*
func Sub(lastNode *sgf.Node, nodesBefore int, opts SubOpts) (*sgf.Node, error) {
	if nodesBefore < 0 {
		nodesBefore = -nodesBefore
	}
	line := lastNode.GetLine()

	// for n, node := range line {
	// 	if w, ok := node.GetValue(SGFTagWhiteMove); ok && w != "" {
	// 		println(n, w)
	// 	} else if b, ok := node.GetValue(SGFTagLabel); ok && b != "" {
	// 		println(n, b)
	// 	}
	// }

	from := len(line) - nodesBefore - 1
	if from < 0 {
		from = 0
	}
	line = line[from:]

	rootNode := sgf.NewNode(nil)
	tmpNode := rootNode
	for n, node := range line {
		if n == 0 {
			CopyKeys(FindMetadataNode(node), tmpNode, SGFTagComment) // To get board size and all other important metadata
			CopyKeys(node, tmpNode, SGFTagBlackMove, SGFTagWhiteMove)

			if w, ok := node.GetValue(SGFTagWhiteMove); ok && w != "" {
				tmpNode.SetValue(SGFTagCircle, w)
			} else if b, ok := node.GetValue(SGFTagBlackMove); ok && b != "" {
				tmpNode.SetValue(SGFTagCircle, b)
			}
			board := node.Board()
			for row := 0; row < board.Size; row++ {
				for col := 0; col < board.Size; col++ {
					color := board.Get(sgf.Point(col, row))
					coord := sgf.Point(col, row)
					if color == sgf.BLACK {
						tmpNode.AddValue(SGFTagAddBlack, coord)
					}
					if color == sgf.WHITE {
						tmpNode.AddValue(SGFTagAddWhite, coord)
					}
				}
			}
		} else {
			var err error
			if w, _ := node.GetValue(SGFTagWhiteMove); w != "" {
				tmpNode, err = tmpNode.PlayColour(w, sgf.WHITE)
				if err != nil {
					return nil, err
				}
			} else if b, _ := node.GetValue(SGFTagBlackMove); b != "" {
				tmpNode, err = tmpNode.PlayColour(b, sgf.BLACK)
				if err != nil {
					return nil, err
				}
			}
			CopyKeys(node, tmpNode, SGFTagBlackMove, SGFTagWhiteMove)
		}
	}
	return rootNode, nil
}

func FindMetadataNode(node *sgf.Node) *sgf.Node {
	for _, n := range node.GetLine() {
		if _, ok := n.GetValue(SGFTagSize); ok {
			return n
		}
	}
	return node
}

func CopyKeys(from, to *sgf.Node, except ...string) {
_loop:
	for _, key := range from.AllKeys() {
		for _, exc := range except {
			if exc == key {
				continue _loop
			}
		}
		to.SetValues(key, from.AllValues(key))
	}
}
