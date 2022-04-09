package main

import (
	"bytes"
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	renameAll := false
files_loop:
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			return err
		}

		gi := sgfutils.ParseGameInfo(node)

		dir := path.Dir(fn)
		newName := path.Join(dir, gi.SuggestedFilename+"-"+getMovesHash(node)+".sgf")
		fmt.Println(fn, "->", newName)

		if fn == newName {
			fmt.Println("OK")
			continue files_loop
		}

		if _, err := os.Stat(newName); err == nil {
			fmt.Fprintf(os.Stderr, "File %s already exists\n", newName)
			os.Exit(1)
		} else {
			var resp string
			if renameAll {
				resp = "a"
			} else {
				resp = strings.ToLower(sgfutils.Input("Rename? [(y)es (N)o (A)all]"))
			}
			renameAll = resp == "a"
			if renameAll || resp == "y" {
				if err := os.Rename(fn, newName); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getMovesHash(node *sgf.Node) string {
	movesCoords := bytes.NewBufferString("")
	walkNodes(node, 0, movesCoords)
	h := md5.New()
	str := movesCoords.String()
	hash := fmt.Sprintf("%x", h.Sum(nil))[0:5]
	//fmt.Println(str, "->", hash)
	io.WriteString(h, str)
	return hash
}

func walkNodes(node *sgf.Node, depth int, moves *bytes.Buffer) {
	for _, child := range node.Children() {
		for _, tag := range []string{sgfutils.SGFTagBlackMove, sgfutils.SGFTagWhiteMove} {
			if value, found := node.GetValue(tag); found && value != "" {
				if value != "" {
					_, _ = moves.WriteString(value)
				}
			}
		}
		walkNodes(child, depth+1, moves)
	}
}
