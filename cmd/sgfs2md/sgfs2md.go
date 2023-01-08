package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/tkrajina/sgf2img/sgfutils/sgf2img"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	b := bytes.NewBufferString("")

	opts := sgf2img.Options{AutoCrop: true, ImageSize: 150, ImageType: sgf2img.SVG, Images: []int{1}}
	for _, fn := range flag.Args() {
		node, images, err := sgf2img.ProcessSgfFile(fn, &opts)
		_ = node
		panicIfErr(err)

		_, _ = b.WriteString(fmt.Sprintf("# %s\n\n", fn))
		if len(images) > 0 {
			_, _ = b.Write(images[0].Contents)
			_, _ = b.WriteString("\n\n")
		}
	}

	os.WriteFile("sgfs.md", b.Bytes(), 0700)
}
