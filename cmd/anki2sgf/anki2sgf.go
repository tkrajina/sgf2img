package main

import (
	"strings"

	"github.com/rooklift/sgf"
)

const board = `...................
..............www..
...w......b.w.bwb..
..............bbb..
...w...............
...................
...................
...............b...
................b..
...................
..............bbw..
...............w.w.
.............bbbww.
..............wwbb.
...w........w.wb...
.........b.....b.w.
...w.........w.b.B.
...................
...................
PW:katago with ELO 10000 custom config
PB:katago with ELO 10000
AP:Sabaki:0.52.0
DT:2022-03-05
RE:B+Resign
FN:sgf/example_game.sgf
TR:pp,qi
SQ:po
CR:qn
MA:pq
LB:rn:A,rp:1
--
17:.................W.

--
16:...w.........w.bBb.

--
17:..............W..w.

--
14:...w........w.wb.B.

--
16:...w......W..w.bbb.

--
2:...w...B..b.w.bwb..`

func main() {
	root := sgf.NewTree(19)
	root.Board().AddStone(sgf.Point(0, 0), sgf.BLACK)
}

func splitMoves(txt string) (board string, moves [][]string) {
	var initialBoard []string
	for _, line := strings.Split(txt, "\n") {
		if len(moves) == 0 {
			moves = append(moves, []string{})
		}
		line = strings.TrimSpace(line)
		lineLower := strings.ToLower(line)
		if len(line) == 0 {
			continue
		} else if strings.HasPrefix(line, "--") {
		} else if lineLower[0] == "b" || lineLower[0] == "w" || lineLower[0] == "." {
		} else {
		}
	}
}
