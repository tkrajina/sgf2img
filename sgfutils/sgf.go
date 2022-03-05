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
