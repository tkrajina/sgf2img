package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

type Opts struct {
	Recursive bool
	GobanTxt  [][]sgf.Colour
	filename  string
}

func main() {
	flag.Parse()

	gobanTxt := inputMultiline("Goban position (empty line to finish input):", func(line string) bool {
		return strings.TrimSpace(line) == ""
	})

	var opts Opts

	for _, line := range strings.Split(gobanTxt, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		opts.GobanTxt = append(opts.GobanTxt, []sgf.Colour{})
		for _, r := range []rune(line) {
			s := string(r)
			if strings.TrimSpace(s) == "" {
				continue
			}
			var color sgf.Colour
			switch strings.ToLower(s) {
			case sgfutils.WhiteCircle:
				color = sgf.WHITE
			case sgfutils.BlackCircle:
				color = sgf.BLACK
			case sgfutils.Empty:
				color = sgf.EMPTY
			default:
				panic("Invalid character: " + s + ".")
			}
			opts.GobanTxt[len(opts.GobanTxt)-1] = append(opts.GobanTxt[len(opts.GobanTxt)-1], color)
		}
	}

	for _, sgfFn := range flag.Args() {
		fnOpts := opts
		fnOpts.filename = sgfFn
		if err := processSgfFile(sgfFn, fnOpts); err != nil {
			panic(err.Error())
		}
	}
}

func processSgfFile(sgfFn string, opts Opts) error {
	fmt.Println("Loading", sgfFn)
	node, err := sgf.Load(sgfFn)
	if err != nil {
		return err
	}

	return walkNodesAndMarkMistakes(node, opts, 0)
}

func gobanEquals(board sgf.Board, opts Opts) bool {
	// for lineNo, line := range opts.GobanTxt {
	// 	for columnNo := range line {
	// 		pos := board.Get(sgf.Point(columnNo, lineNo))
	// 		fmt.Print(pos)
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Println()

	// for _, line := range opts.GobanTxt {
	// 	for _, gobanPos := range line {
	// 		fmt.Print(gobanPos)
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Println()

	for lineNo, line := range opts.GobanTxt {
		for columnNo, gobanPos := range line {
			pos := board.Get(sgf.Point(columnNo, lineNo))
			if gobanPos != pos {
				return false
			}
		}
	}
	return true
}

func walkNodesAndMarkMistakes(node *sgf.Node, opts Opts, depth int) error {
	if gobanEquals(*node.Board(), opts) {
		fmt.Println("Found:")
		fmt.Println(sgfutils.BoardToString(*node.Board()))
		fmt.Println("File: " + opts.filename)
		fmt.Printf("Move: %d\n", depth+1)
		return nil
	}

	for _, child := range node.Children() {
		if err := walkNodesAndMarkMistakes(child, opts, depth+1); err != nil {
			return err
		}
	}

	return nil
}

func inputMultiline(msg string, lastLine func(line string) bool) (txt string) {
	scn := bufio.NewScanner(os.Stdin)
	fmt.Println(msg)
	var lines []string
	for scn.Scan() {
		line := scn.Text()
		txt = strings.Join(lines, "\n")
		if lastLine(line) {
			return
		}
		lines = append(lines, line)
	}

	return
}
