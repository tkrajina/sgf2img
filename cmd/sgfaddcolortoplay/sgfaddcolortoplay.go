package main

import (
	"flag"
	"fmt"
	"os"

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
	for _, fn := range flag.Args() {
		fmt.Println("Loading", fn)
		s, err := sgf.Load(fn)
		panicIfErr(err)
		color, added := sgfutils.AddMissingPLayerColor(s)
		if added {
			fmt.Println("  Added color to play:", color)
		} else {
			fmt.Println("  Nothing to do")
		}
		err = os.WriteFile(fn, []byte(s.SGF()), 0700)
		panicIfErr(err)
	}
}
