package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	var nodeNumbers string
	flag.Int64Var(&(opts.ImageSize), "s", 400, "Image size (max goban board image size)")
	flag.BoolVar(&(opts.Grayscale), "g", false, "Grayscale only for png images")
	flag.BoolVar(&(opts.MainLine), "ml", false, "Make one image out of the main branch line")
	flag.BoolVar(&(opts.Verbose), "v", false, "Verbose")
	flag.BoolVar(&(opts.AutoCrop), "c", false, "Autocrop")
	flag.StringVar(&nodeNumbers, "n", "0", "Node numbers (coma separated, -1 for last node)")
	flag.StringVar(&typ, "t", string(sgf2img.PNG), fmt.Sprintf("Image type (%s|%s)", sgf2img.PNG, sgf2img.SVG))
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	for _, nStr := range strings.Split(nodeNumbers, ",") {
		nStr = strings.TrimSpace(nStr)
		n, err := strconv.ParseInt(nStr, 10, 32)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid node numbers: "+nodeNumbers)
			os.Exit(1)
		}
		opts.Images = append(opts.Images, int(n))
	}

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
		_, files, err := sgf2img.ProcessSgfFile(sgfFn, &opts)
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
