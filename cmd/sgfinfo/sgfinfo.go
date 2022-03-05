package main

import (
	"flag"
	"fmt"

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
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			return err
		}

		gi := sgfutils.ParseGameInfo(node)
		fmt.Printf("%20s %5s vs %20s %5s: %s\n", gi.BlackName, gi.BlackRank, gi.WhiteName, gi.WhiteRank, fn)
	}
	return nil
}
