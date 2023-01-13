package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Opts struct {
	filename string
}

var recursive bool

func main() {
	flag.BoolVar(&recursive, "r", false, "Recursive")
	flag.Parse()

	gobanTxt := inputMultiline("SGF (empty line to finish input):", func(line string) bool {
		return strings.TrimSpace(line) == ""
	})

	fn := path.Join(os.TempDir(), fmt.Sprintf("tmp_%d.sgf", time.Now().Unix()))
	defer os.Remove(fn)

	panicIfErr(ioutil.WriteFile(fn, []byte(gobanTxt), 0700))

	s, err := sgf.Load(fn)
	panicIfErr(err)

	lastNode := s
	for lastNode.Children() != nil {
		if len(lastNode.Children()) == 0 {
			break
		}
		lastNode = lastNode.Children()[0]
	}

	for _, sgfFn := range flag.Args() {
		if recursive {
			filepath.Walk(sgfFn, func(fn string, info fs.FileInfo, err error) error {
				exp := strings.ToLower(path.Ext(fn))
				if !info.IsDir() && exp == ".sgf" {
					if err := processSgfFile(fn, lastNode); err != nil {
						panic(err.Error())
					}
				}
				return nil
			})
		} else {
			if err := processSgfFile(sgfFn, lastNode); err != nil {
				panic(err.Error())
			}
		}
	}
}

func processSgfFile(sgfFn string, targetNode *sgf.Node) error {
	//fmt.Println("Loading", sgfFn)
	node, err := sgf.Load(sgfFn)
	if err != nil {
		return err
	}

	return walkNodes(sgfFn, node, targetNode, 0)
}

func gobanEquals(board sgf.Board, node *sgf.Node) bool {
	nodeBoard := node.Board()
	if board.Size != nodeBoard.Size {
		return false
	}
	for lineNo := 0; lineNo < board.Size; lineNo++ {
		for columnNo := 0; columnNo < board.Size; columnNo++ {
			if board.Get(sgf.Point(columnNo, lineNo)) != nodeBoard.Get(sgf.Point(columnNo, lineNo)) {
				return false
			}
		}
	}
	return true
}

func walkNodes(filename string, node *sgf.Node, targetNode *sgf.Node, depth int) error {
	if gobanEquals(*node.Board(), targetNode) {
		fmt.Println("Found:")
		fmt.Println(sgfutils.BoardToString(*node.Board()))
		fmt.Println("File: " + filename)
		fmt.Printf("Move: %d\n", depth)
		return nil
	}

	for _, child := range node.Children() {
		if err := walkNodes(filename, child, targetNode, depth+1); err != nil {
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
