package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
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

var (
	diffLine = regexp.MustCompile(`^\d+:[\.wbWB]+`)
	tagLine  = regexp.MustCompile(`^\w+:[\.wbWB]+`)
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	node := sgf.NewTree(19)
	node.Board().AddStone(sgf.Point(0, 0), sgf.BLACK)

	var err error
	_ = err
	var initialBoard []string
	moves := splitMoves(board)
	for n, move := range moves {
		fmt.Println("Move:\n" + strings.Join(move, "\n") + "\n")
		for _, line := range move {
			lineLower := strings.ToLower(line)
			if n == 0 && (lineLower[0] == '.' || lineLower[0] == 'b' || lineLower[0] == 'w') {
				fmt.Println("initial:", line)
				initialBoard = append(initialBoard, line)
			} else if n > 0 && diffLine.MatchString(line) {
				parts := strings.Split(line, ":")
				lineNo, err := strconv.ParseInt(parts[0], 10, 16)
				_ = lineNo
				panicIfErr(err)
				whiteColumn := strings.IndexRune(parts[1], 'W') // TODO
				blackColumn := strings.IndexRune(parts[1], 'B') // TODO
				if whiteColumn >= 0 {
					//node, err = node.PlayColour(sgf.Point(whiteColumn, int(lineNo)), sgf.WHITE)
					//panicIfErr(err)
				} else if blackColumn >= 0 {
					//node, err = node.PlayColour(sgf.Point(blackColumn, int(lineNo)), sgf.BLACK)
					//panicIfErr(err)
				}
			} else if tagLine.MatchString(line) {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					node.AddValue(parts[0], parts[1])
				}
			}
		}
		if n == 0 {
			for lineNo, line := range initialBoard {
				for columNo, rune := range []rune(line) {
					coords := sgf.Point(columNo, lineNo)
					switch rune {
					case 'b':
						node.AddValue(sgfutils.SGFTagAddBlack, coords)
					case 'B':
						_, err := node.PlayColour(coords, sgf.BLACK)
						panicIfErr(err)
					case 'w':
						node.AddValue(sgfutils.SGFTagAddWhite, coords)
					case 'W':
						_, err := node.PlayColour(coords, sgf.WHITE)
						panicIfErr(err)
					}
				}
			}
		}
	}

	node.Save("tmp.sgf")
}

func splitMoves(txt string) (moves [][]string) {
	for _, line := range strings.Split(txt, "\n") {
		line = strings.TrimSpace(line)
		if len(moves) == 0 {
			moves = append(moves, []string{})
		}
		if strings.HasPrefix(line, "--") {
			moves = append(moves, []string{})
		} else if line != "" {
			moves[len(moves)-1] = append(moves[len(moves)-1], line)
		}
	}
	return
}
