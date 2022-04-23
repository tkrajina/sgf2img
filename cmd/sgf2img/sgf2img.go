package main

import (
	"encoding/csv"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dsvg"

	"github.com/kettek/apng"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type imageType string

const (
	png  imageType = "png"
	svg  imageType = "svg"
	html imageType = "html"
	anki imageType = "anki"
)

type ctx struct {
	imageSize  int64
	imageType  imageType
	ankiImport string
	grayscale  bool
	mistakes   bool
	verbose    bool
	csvRows    [][]string
}

func main() {
	var opts ctx
	var help bool
	var typ string
	flag.Int64Var(&(opts.imageSize), "s", 400, "Image size (max goban board image size)")
	flag.StringVar(&(opts.ankiImport), "a", "", "Create Anki import file")
	flag.BoolVar(&(opts.grayscale), "g", false, "Grayscale only for png images")
	flag.BoolVar(&(opts.mistakes), "m", false, "Mistakes to images (assumes that if a node comment starts with 'Mistake' the parent has another branch which is the right path)")
	flag.BoolVar(&(opts.verbose), "v", false, "Verbose")
	flag.StringVar(&typ, "t", string(png), fmt.Sprintf("Image type (%s|%s|%s|%s)", png, svg, html, anki))
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()

	opts.imageType = imageType(typ)

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Println("No SGF files given")
		os.Exit(1)
	}

	for _, sgfFn := range flag.Args() {
		if err := processSgfFile(sgfFn, &opts); err != nil {
			panic(err.Error())
		}
	}

	if len(opts.csvRows) > 0 {
		cssFn := exportedImgFilename("anki_import", "", "", "csv")
		f, err := os.Create(cssFn)
		panicIfErr(err)
		csvwriter := csv.NewWriter(f)
		panicIfErr(csvwriter.WriteAll(opts.csvRows))
		fmt.Printf("Saved %d rows to %s\n", len(opts.csvRows), cssFn)
	}
}

func processSgfFile(sgfFn string, opts *ctx) error {
	fmt.Println("Loading", sgfFn)
	node, err := sgf.Load(sgfFn)
	if err != nil {
		return err
	}

	if opts.mistakes {
		walkNodesAndMarkMistakes(node, opts, 0)
		fn := "tmp.sgf" //path.Join(os.TempDir(), "tmp.sgf")
		if err := node.Save(fn); err != nil {
			return err
		}
		fmt.Println("Saved marked file to", fn)
	}

	if err := walkNodes(sgfFn, node, opts, 0); err != nil {
		return err
	}

	return nil
}

func animatePng(images []image.Image, fn string) error {
	a := apng.APNG{
		Frames:    make([]apng.Frame, len(images)),
		LoopCount: 1,
	}
	// Open our file for writing
	out, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer out.Close()
	// Assign each decoded PNG's Image to the appropriate Frame Image
	for n := range images {
		a.Frames[n].Image = images[n]
		if n < len(images)-1 {
			a.Frames[n].DelayNumerator = 1
		} else {
			a.Frames[n].DelayNumerator = 6
		}
		a.Frames[n].DelayDenominator = 2
	}
	// Write APNG to our output file
	return apng.Encode(out, a)
}

func exportedImgFilename(sgfFn, name, suffix, extension string) string {
	dir, file := path.Split(sgfFn)
	base := strings.Replace(file, path.Ext(file), "", 1)
	return path.Join(dir, strings.Trim("sgf2img_goban_"+base+"_"+name+"_"+suffix, "_")+"."+extension)
}

func walkNodes(sgfFilename string, node *sgf.Node, opts *ctx, depth int) error {
	comment := parseNodeComment(node)
	for _, ci := range comment.images {
		if opts.verbose {
			fmt.Println(sgfutils.BoardToString(*node.Board()))
		}

		////////////////////////////////////////////////////////////////////////////////////////////////////
		////////////////////////////////////////////////////////////////////////////////////////////////////
		fn := exportedImgFilename(sgfFilename, ci.name, "", string(opts.imageType))
		switch opts.imageType {
		case svg:
			svg := draw2dsvg.NewSvg()
			boardToImage(draw2dsvg.NewGraphicContext(svg), *node, *opts)
			byts, err := xml.Marshal(svg)
			if err != nil {
				return err
			}
			ioutil.WriteFile(fn, byts, 0700)
		case png:
			dest := image.NewRGBA(image.Rect(0, 0, int(opts.imageSize), int(opts.imageSize)))
			gc := draw2dimg.NewGraphicContext(dest)
			boardToImage(gc, *node, *opts)
			// img = crop(img, ci, *node.Board(), opts) TODO

			if opts.grayscale {
				gs := grayscale(dest, *opts)
				draw2dimg.SaveToPngFile(fn, crop(gs, ci, *node.Board(), *opts))
			} else {
				draw2dimg.SaveToPngFile(fn, crop(dest, ci, *node.Board(), *opts))
			}
		case anki:
			txt := boardToAnki(*node) + "\n"
			txt += boardAnnotationsToAnki(*node.GetRoot())
			txt += fmt.Sprintf("FN:%s\n", sgfFilename)
			txt += nodeAnnotationsToAnki(*node)
			if err := ioutil.WriteFile(fn, []byte(txt), 0700); err != nil {
				return err
			}
		case html:
			// TODO
		default:
			return fmt.Errorf("invalid type: %s", opts.imageType)
		}
		fmt.Printf("Saved 1 board position on move %d (%s) to: %s\n", depth, ci.name, fn)
	}

	if err := saveAnimations(comment, node, opts, sgfFilename, depth); err != nil {
		return err
	}

	// Continue recursion
	for _, child := range node.Children() {
		if err := walkNodes(sgfFilename, child, opts, depth+1); err != nil {
			return err
		}
	}

	return nil
}

func grayscale(dest image.Image, opts ctx) image.Image {
	w, h := int(opts.imageSize), int(opts.imageSize)
	grayScale := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			imageColor := dest.At(x, y)
			rr, gg, bb, _ := imageColor.RGBA()
			r := math.Pow(float64(rr), 2.2)
			g := math.Pow(float64(gg), 2.2)
			b := math.Pow(float64(bb), 2.2)
			m := math.Pow(0.2125*r+0.7154*g+0.0721*b, 1/2.2)
			Y := uint16(m + 0.5)
			grayColor := color.Gray{uint8(Y >> 8)}
			grayScale.Set(x, y, grayColor)
		}
	}
	return grayScale
}

func saveAnimations(cm commentMedatada, node *sgf.Node, opts *ctx, sgfFilename string, depth int) error {
	for _, ca := range cm.animate {
		tmpNode := node
		var parentImage commentImage

		animatedNodes := []sgf.Node{*tmpNode}

	loop:
		for true {
			parentCm := parseNodeComment(tmpNode)
			for _, parentCi := range parentCm.start {
				if parentCi.name == ca.name {
					parentImage = parentCi
					break loop
				}
			}
			tmpNode = tmpNode.Parent()
			if tmpNode == nil {
				if opts.verbose {
					fmt.Println(sgfutils.BoardToString(*node.Board()))
				}
				return fmt.Errorf("can't find node with img name '%s' for animation (loc %d)", ca.name, depth+1)
			} else {
				animatedNodes = append([]sgf.Node{*tmpNode}, animatedNodes...)
			}
		}

		fn := exportedImgFilename(sgfFilename, ca.name, "animated", string(opts.imageType))
		switch opts.imageType {
		case png:
			var images []image.Image
			for _, n := range animatedNodes {
				img := image.NewRGBA(image.Rect(0, 0, int(opts.imageSize), int(opts.imageSize)))
				gc := draw2dimg.NewGraphicContext(img)
				boardToImage(gc, n, *opts)
				images = append(images, img)
			}
			fmt.Printf("Found %d images to animate\n", len(images))
			for n := range images {
				if opts.grayscale {
					images[n] = grayscale(images[n], *opts)
				}
				images[n] = crop(images[n], parentImage, *node.Board(), *opts)
			}
			if err := animatePng(images, fn); err != nil {
				return err
			}
		case anki:
			txt := ""
			for n, node := range animatedNodes {
				if n > 0 {
					txt += "--" + "\n"
				}
				if n == 0 {
					txt += boardToAnki(node) + "\n"
					txt += boardAnnotationsToAnki(*node.GetRoot())
					txt += fmt.Sprintf("FN:%s\n", sgfFilename)
					txt += nodeAnnotationsToAnki(node)
				} else {
					prev := strings.Split(boardToAnki(animatedNodes[n-1]), "\n")
					this := strings.Split(boardToAnki(node), "\n")
					for n := 0; n < len(prev); n++ {
						if strings.ToLower(prev[n]) != strings.ToLower(this[n]) {
							txt += fmt.Sprintf("%d:%s\n", n, strings.TrimSpace(this[n]))
						}
					}
					txt += nodeAnnotationsToAnki(node)
				}
			}
			ioutil.WriteFile(fn, []byte(txt), 0700)
			opts.csvRows = append(opts.csvRows, []string{"<pre>" + txt + "</pre>"})
		case html:
			// TODO
		}
		fmt.Printf("Saved %d board positions on move %d (%s) to: %s\n", len(animatedNodes), depth, ca.name, fn)
	}

	return nil
}

func crop(img image.Image, cm commentImage, board sgf.Board, opts ctx) image.Image {
	band := float64(opts.imageSize) / float64(board.Size)
	left := float64(cm.left) * band
	right := float64(cm.right) * band
	up := float64(cm.up) * band
	down := float64(cm.down) * band
	if left+right > float64(opts.imageSize) {
		left = float64(opts.imageSize) / 2.
		right = float64(opts.imageSize) / 2.
	}
	if up+down > float64(opts.imageSize) {
		up = float64(opts.imageSize) / 2.
		down = float64(opts.imageSize) / 2.
	}
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(int(left), int(up), int(opts.imageSize)-int(right), int(opts.imageSize)-int(down)))
}

func sgfCoordinatesToImageCoordinates(coords string, imagesize int, board sgf.Board) (float64, float64) {
	x, y, _ := sgf.ParsePoint(coords, board.Size)
	return boardCoordinateToImageCoordinate(x, y, imagesize, board)
}

func boardCoordinateToImageCoordinate(boardX, boardY, imagesize int, board sgf.Board) (float64, float64) {
	band := float64(imagesize) / float64(board.Size)
	return float64(boardX)*band + float64(band/2), float64(boardY)*band + float64(band/2)
}

// svg: https://github.com/ajstarks/svgo
// generate animated gif

// draw2d https://github.com/llgcode/draw2d

// https://github.com/kettek/apngr animated png
