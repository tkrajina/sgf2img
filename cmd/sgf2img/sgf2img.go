package main

import (
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
	var opts sgf2img.Options
	var help bool
	var typ string
	flag.Int64Var(&(opts.ImageSize), "s", 400, "Image size (max goban board image size)")
	flag.StringVar(&(opts.AnkiImport), "a", "", "Create Anki import file")
	flag.BoolVar(&(opts.Grayscale), "g", false, "Grayscale only for png images")
	flag.BoolVar(&(opts.Mistakes), "mi", false, "Mistakes to images (assumes that if a node comment starts with 'Mistake' the parent has another branch which is the right path)")
	flag.BoolVar(&(opts.MainLine), "ml", false, "Make one image out of the main branch line")
	flag.BoolVar(&(opts.Verbose), "v", false, "Verbose")
	flag.StringVar(&typ, "t", string(sgf2img.PNG), fmt.Sprintf("Image type (%s|%s)", sgf2img.PNG, sgf2img.SVG))
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	if typ == "" {
		opts.ImageType = sgf2img.PNG
	} else {
		opts.ImageType = sgf2img.ImageType(typ)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("No SGF files given")
		os.Exit(1)
	}

	for _, sgfFn := range flag.Args() {
		files, err := sgf2img.ProcessSgfFile(sgfFn, &opts)
		if err != nil {
			panic(err.Error())
		}
		for _, file := range files {
			fmt.Printf("Save %d bytes in %s\n", len(file.Contents), file.Name)
			if err := os.WriteFile(file.Name, file.Contents, 0700); err != nil {
				panic(err.Error())
			}
		}
	}
}
