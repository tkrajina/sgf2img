package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

var (
	comments bool
	branches bool
	outfile  string
	all      bool
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.BoolVar(&comments, "c", false, "Strip comments")
	flag.BoolVar(&branches, "b", false, "Strip branches")
	flag.BoolVar(&all, "a", false, "Strip all")
	flag.StringVar(&outfile, "o", "", "Output file")
	flag.Parse()

	if all {
		comments = true
		branches = true
	}

	if !comments && !branches {
		flag.Usage()
		return
	}

	for _, fn := range flag.Args() {
		s, err := sgf.Load(fn)
		panicIfErr(err)

		panicIfErr(doStuff(s))

		outFn := outfile
		if outFn == "" {
			ext := filepath.Ext(fn)
			outFn = fn[0:len(fn)-len(ext)] + "_stripped.sgf"
		}

		panicIfErr(s.Save(outFn))
		fmt.Println("Saved to", outFn)
	}
}

func doStuff(node *sgf.Node) error {
	if comments {
		node.DeleteKey(sgfutils.SGFTagComment)
	}
	for n, child := range node.Children() {
		if branches && n > 0 {
			fmt.Println("Remove branch")
			child.Detach()
			continue
		}
		if err := doStuff(child); err != nil {
			return err
		}
	}
	return nil
}
