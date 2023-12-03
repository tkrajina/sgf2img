package sgfutils

import (
	"strings"

	"github.com/rooklift/sgf"
)

func BoardToString(board sgf.Board) string {
	black := "●"
	white := "○"
	var res []string

	for i := 0; i < board.Size; i++ {
		res = append(res, "")
		for j := 0; j < board.Size; j++ {
			color := board.Get(sgf.Point(j, i))
			switch color {
			// https://en.wikipedia.org/wiki/Box-drawing_character
			case sgf.EMPTY:
				ch := "┼─"
				if i == 0 {
					if j == 0 {
						ch = "┌─"
					} else if j == board.Size-1 {
						ch = "┐"
					} else {
						ch = "┬─"
					}
				} else if i == board.Size-1 {
					if j == 0 {
						ch = "└─"
					} else if j == board.Size-1 {
						ch = "┘"
					} else {
						ch = "┴─"
					}
				} else if j == 0 {
					ch = "├─"
				} else if j == board.Size-1 {
					ch = "┤ "
				}
				res[len(res)-1] += ch
			case sgf.WHITE:
				if j == board.Size-1 {
					res[len(res)-1] += string(white)
				} else {
					res[len(res)-1] += string(white) + "─"
				}
			case sgf.BLACK:
				if j == board.Size-1 {
					res[len(res)-1] += string(black)
				} else {
					res[len(res)-1] += string(black) + "─"
				}
			}
		}
	}
	return strings.Join(res, "\n")
}

func FindFirstMove(node *sgf.Node) (color, player, coords string) {
	whiteName, blackName := "", ""
	tmpNode := node
	for len(tmpNode.Children()) > 0 {
		if name, ok := tmpNode.GetValue(SGFTagWhiteName); ok && name != "" {
			whiteName = name
		}
		if name, ok := tmpNode.GetValue(SGFTagBlackName); ok && name != "" {
			blackName = name
		}
		tmpNode = tmpNode.Children()[0]
	}

	tmpNode = node
	for len(tmpNode.Children()) > 0 {
		if coord, ok := tmpNode.GetValue(SGFTagWhiteMove); ok && coord != "" {
			color = "w"
			coords = coord
			player = whiteName
			return
		}
		if coord, ok := tmpNode.GetValue(SGFTagBlackMove); ok && coord != "" {
			color = "b"
			coords = coord
			player = blackName
			return
		}
		tmpNode = tmpNode.Children()[0]
	}
	return
}

func AddMissingPLayerColor(s *sgf.Node) (color string, added bool) {
	color, _, _ = FindFirstMove(s)
	toPlay := s.AllValues(SGFTagPlayer)
	if len(toPlay) == 0 && color != "" {
		s.AddValue(SGFTagPlayer, strings.ToUpper(color))
		added = true
	}
	return
}
