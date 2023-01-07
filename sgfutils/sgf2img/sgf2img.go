package sgf2img

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	imagepng "image/png"
	"math"
	"os"
	"path"
	"strings"

	"github.com/kettek/apng"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dsvg"
	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

type ImageType string

const (
	PNG ImageType = "png"
	SVG ImageType = "svg"
)

type Options struct {
	ImageSize  int64
	ImageType  ImageType
	AnkiImport string
	Grayscale  bool
	Mistakes   bool
	MainLine   bool
	Verbose    bool
}

type GobanImageFile struct {
	Name     string
	Contents []byte
}

func ProcessSgfFile(sgfFn string, opts *Options) (*sgf.Node, []GobanImageFile, error) {
	fmt.Println("Loading", sgfFn)
	node, err := sgf.Load(sgfFn)
	if err != nil {
		return nil, nil, err
	}

	if opts.Mistakes {
		walkNodesAndMarkMistakes(node, opts, 0)
	}
	if opts.MainLine {
		tmpNode := node
		tmpNode.SetValue(directiveStart, "main_line")

		for {
			if len(tmpNode.Children()) > 0 {
				tmpNode = tmpNode.Children()[0]
			} else {
				tmpNode.SetValue(directiveEnd, "main_line")
				break
			}
		}
	}

	files, err := walkNodes(sgfFn, node, opts, 0)
	return node, files, err
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
			a.Frames[n].DelayDenominator = 20
		} else {
			a.Frames[n].DelayNumerator = 6
			a.Frames[n].DelayDenominator = 2
		}
	}
	// Write APNG to our output file
	return apng.Encode(out, a)
}

func exportedImgFilename(sgfFn, name, suffix, extension string) string {
	dir, file := path.Split(sgfFn)
	base := strings.Replace(file, path.Ext(file), "", 1)
	return path.Join(dir, strings.Trim("sgf2img_"+base+"_"+name+"_"+suffix, "_")+"."+extension)
}

func walkNodes(sgfFilename string, node *sgf.Node, opts *Options, depth int) ([]GobanImageFile, error) {
	var files []GobanImageFile

	comment := parseNodeImgMetadata(node)
	for _, ci := range comment.images {
		if opts.Verbose {
			fmt.Println(sgfutils.BoardToString(*node.Board()))
		}

		////////////////////////////////////////////////////////////////////////////////////////////////////
		////////////////////////////////////////////////////////////////////////////////////////////////////
		fn := exportedImgFilename(sgfFilename, ci.name, "", string(opts.ImageType))
		switch opts.ImageType {
		case SVG:
			svg := draw2dsvg.NewSvg()
			boardToImage(draw2dsvg.NewGraphicContext(svg), *node, *opts)
			byts, err := xml.Marshal(svg)
			if err != nil {
				return nil, err
			}
			files = append(files, GobanImageFile{Name: fn, Contents: byts})
		case PNG:
			dest := image.NewRGBA(image.Rect(0, 0, int(opts.ImageSize), int(opts.ImageSize)))
			gc := draw2dimg.NewGraphicContext(dest)
			boardToImage(gc, *node, *opts)
			// img = crop(img, ci, *node.Board(), opts) TODO

			var i image.Image
			if opts.Grayscale {
				gs := grayscale(dest, *opts)
				i = crop(gs, ci, *node.Board(), *opts)
			} else {
				i = crop(dest, ci, *node.Board(), *opts)
			}
			b := bytes.NewBuffer([]byte{})
			if err := imagepng.Encode(b, i); err != nil {
				return nil, err
			}

			files = append(files, GobanImageFile{Name: fn, Contents: b.Bytes()})
		default:
			return nil, fmt.Errorf("invalid type: %s", opts.ImageType)
		}
		fmt.Printf("Saved 1 board position on move %d (%s) to: %s\n", depth, ci.name, fn)
	}

	if err := saveAnimations(comment, node, opts, sgfFilename, depth); err != nil {
		return nil, err
	}

	// Continue recursion
	for _, child := range node.Children() {
		newFiles, err := walkNodes(sgfFilename, child, opts, depth+1)
		if err != nil {
			return nil, err
		}
		files = append(files, newFiles...)
	}

	return files, nil
}

func grayscale(dest image.Image, opts Options) image.Image {
	w, h := int(opts.ImageSize), int(opts.ImageSize)
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

func saveAnimations(cm nodeImgMetdata, node *sgf.Node, opts *Options, sgfFilename string, depth int) error {
	for _, ca := range cm.animate {
		tmpNode := node
		var parentImage commentImage

		animatedNodes := []sgf.Node{*tmpNode}

	loop:
		for true {
			parentCm := parseNodeImgMetadata(tmpNode)
			for _, parentCi := range parentCm.start {
				if parentCi.name == ca.name {
					parentImage = parentCi
					break loop
				}
			}
			tmpNode = tmpNode.Parent()
			if tmpNode == nil {
				if opts.Verbose {
					fmt.Println(sgfutils.BoardToString(*node.Board()))
				}
				return fmt.Errorf("can't find node with img name '%s' for animation (loc %d)", ca.name, depth+1)
			} else {
				animatedNodes = append([]sgf.Node{*tmpNode}, animatedNodes...)
			}
		}

		fn := exportedImgFilename(sgfFilename, ca.name, "animated", string(opts.ImageType))
		switch opts.ImageType {
		case PNG:
			var images []image.Image
			for _, n := range animatedNodes {
				img := image.NewRGBA(image.Rect(0, 0, int(opts.ImageSize), int(opts.ImageSize)))
				gc := draw2dimg.NewGraphicContext(img)
				boardToImage(gc, n, *opts)
				images = append(images, img)
			}
			fmt.Printf("Found %d images to animate\n", len(images))
			for n := range images {
				if opts.Grayscale {
					images[n] = grayscale(images[n], *opts)
				}
				images[n] = crop(images[n], parentImage, *node.Board(), *opts)
			}
			if err := animatePng(images, fn); err != nil {
				return err
			}
		}
		fmt.Printf("Saved %d board positions on move %d (%s) to: %s\n", len(animatedNodes), depth, ca.name, fn)
	}

	return nil
}

func autocropAnki(nodes []sgf.Node) string {
	var bm boardMargins
	for n := range nodes {
		if n == 0 {
			bm = margins(*nodes[n].Board())
		} else {
			bm.add(margins(*nodes[n].Board()))
		}
	}
	//size := float64(nodes[0].Board().Size)
	//fmt.Printf("%#v\n", bm)
	return bm.cropLine(nodes[0].Board().Size)
	//return fmt.Sprintf("CROP:%.2f %.2f %.2f %.2f\n", float64(bm.top)/float64(size), float64(bm.left)/size, float64(bm.bottom)/size, float64(bm.right)/size)
}

type boardMargins struct{ top, right, bottom, left int }

func (gm *boardMargins) add(bm2 boardMargins) {

}

func (bm boardMargins) cropValue(lines, size int) string {
	if lines < 6 {
		return "0" // If close to border => no crop
	}
	val := float64(lines-4) / float64(size)
	if val <= 0 {
		return "0"
	}
	return fmt.Sprintf("%.2f", val)
}

func (bm boardMargins) cropLine(size int) string {
	return fmt.Sprintf("CROP:%s %s %s %s\n",
		bm.cropValue(bm.top, size),
		bm.cropValue(bm.right, size),
		bm.cropValue(bm.bottom, size),
		bm.cropValue(bm.left, size),
	)
}

func margins(b sgf.Board) boardMargins {
	res := boardMargins{0, 0, 0, 0}

	positions := [][]sgf.Colour{}
	for x := 0; x < b.Size; x++ {
		positions = append(positions, []sgf.Colour{})
		for y := 0; y < b.Size; y++ {
			pos := b.Get(sgf.Point(y, x))
			positions[len(positions)-1] = append(positions[len(positions)-1], pos)
		}
		//fmt.Println(positions[len(positions)-1])
	}

top_loop:
	for i := 0; i < b.Size; i++ {
		for j := 0; j < b.Size; j++ {
			if positions[i][j] != sgf.EMPTY {
				res.top = i
				break top_loop
			}
		}
	}

bottom_loop:
	for i := 0; i < b.Size; i++ {
		for j := 0; j < b.Size; j++ {
			if positions[b.Size-i-1][j] != sgf.EMPTY {
				res.bottom = i
				break bottom_loop
			}
		}
	}

right_loop:
	for i := 0; i < b.Size; i++ {
		for j := 0; j < b.Size; j++ {
			if positions[j][b.Size-i-1] != sgf.EMPTY {
				res.right = i
				break right_loop
			}
		}
	}

left_loop:
	for i := 0; i < b.Size; i++ {
		for j := 0; j < b.Size; j++ {
			if positions[j][i] != sgf.EMPTY {
				res.left = i
				break left_loop
			}
		}
	}

	return res
}

func crop(img image.Image, cm commentImage, board sgf.Board, opts Options) image.Image {
	band := float64(opts.ImageSize) / float64(board.Size)
	left := float64(cm.left) * band
	right := float64(cm.right) * band
	up := float64(cm.up) * band
	down := float64(cm.down) * band
	if left+right > float64(opts.ImageSize) {
		left = float64(opts.ImageSize) / 2.
		right = float64(opts.ImageSize) / 2.
	}
	if up+down > float64(opts.ImageSize) {
		up = float64(opts.ImageSize) / 2.
		down = float64(opts.ImageSize) / 2.
	}
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(int(left), int(up), int(opts.ImageSize)-int(right), int(opts.ImageSize)-int(down)))
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
