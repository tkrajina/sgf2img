package main

import (
	"flag"
	"fmt"
	"path/filepath"

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
		s, err := sgf.Load(fn)
		panicIfErr(err)

		panicIfErr(sgfutils.CleanKatrainStuff(s))

		ext := filepath.Ext(fn)
		fn := fn[0:len(fn)-len(ext)] + "_cleaned.sgf"
		panicIfErr(s.Save(fn))
		fmt.Println("Saved to", fn)
	}
}
