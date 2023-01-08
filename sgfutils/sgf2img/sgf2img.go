package sgf2img

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	imagepng "image/png"
	"math"
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
	Images     []int
	ImageType  ImageType
	AnkiImport string
	Grayscale  bool
	Mistakes   bool
	MainLine   bool
	Verbose    bool
	AutoCrop   bool
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
	for _, imgNo := range opts.Images {
		tmpNode := node
		n := 0
		for tmpNode.MainChild() != nil {
			if n == imgNo {
				tmpNode.SetValue(directiveImg, fmt.Sprintf("_img_%d", imgNo))
			}
			tmpNode = tmpNode.MainChild()
			n++
		}
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

func animatePng(images []image.Image, fn string) ([]byte, error) {
	a := apng.APNG{
		Frames:    make([]apng.Frame, len(images)),
		LoopCount: 1,
	}
	// Open our file for writing
	out := bytes.NewBuffer([]byte{})
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
	if err := apng.Encode(out, a); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func emptyLinesAround(node *sgf.Node) (up, down, left, right int) {
	fmt.Println(sgfutils.BoardToString(*node.Board()))
	if len(node.AllValues(sgfutils.SGFTagBlackMove)) == 0 && len(node.AllValues(sgfutils.SGFTagWhiteMove)) == 0 {
		node = node.MainChild()
	}
	if node == nil {
		return
	}
	var (
		board  = node.Board()
		size   = board.Size
		rowMax = 0
		rowMin = size
		colMax = 0
		colMin = size
	)
	stones := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			color := board.Get(sgf.Point(col, row))
			if color == sgf.BLACK || color == sgf.WHITE {
				stones++
				if row > rowMax {
					rowMax = row
				}
				if col > colMax {
					colMax = col
				}
				if row < rowMin {
					rowMin = row
				}
				if col < colMin {
					colMin = col
				}
			}
		}
	}
	if stones == 0 {
		return 0, 0, 0, 0
	}
	return rowMin, size - rowMax - 1, colMin, size - colMax - 1
}

func calculateCrop(nodes []*sgf.Node, opts Options) (crop Crop, originalImgSize int) {
	originalImgSize = int(opts.ImageSize)
	if !opts.AutoCrop {
		return
	}
	boardSize := -1
	var up, down, left, right int
	for n, node := range nodes {
		if n == 0 {
			up, down, left, right = emptyLinesAround(node)
			boardSize = node.Board().Size
		} else {
			u, d, l, r := emptyLinesAround(node)
			if u < up {
				up = u
			}
			if d < down {
				down = d
			}
			if l < left {
				left = l
			}
			if r < right {
				right = r
			}
		}
		if opts.Verbose {
			fmt.Printf("node #%d -- empty lines: %d %d %d %d\n", n, up, down, left, right)
		}
	}
	if opts.Verbose {
		fmt.Printf("crop: %d %d %d %d\n", up, down, left, right)
	}
	crop.Up = up
	crop.Down = down
	crop.Left = left
	crop.Right = right
	crop.Bigger(2)

	resize := float64(boardSize) / math.Max(
		float64(boardSize-crop.Left-crop.Right),
		float64(boardSize-crop.Up-crop.Down),
	)
	originalImgSize = int(float64(opts.ImageSize) * resize)
	if originalImgSize <= 0 {
		originalImgSize = int(opts.ImageSize)
	}
	fmt.Println("resize:", resize, "to", originalImgSize)

	return
}

func exportedImgFilename(sgfFn, name, suffix, extension string) string {
	dir, file := path.Split(sgfFn)
	base := strings.Replace(file, path.Ext(file), "", 1)
	return path.Join(dir, strings.Trim("sgf2img_"+base+"_"+name+"_"+suffix, "_")+"."+extension)
}

func walkNodes(sgfFilename string, node *sgf.Node, opts *Options, depth int) ([]GobanImageFile, error) {
	var files []GobanImageFile
	boardSize := node.Board().Size

	comment := parseNodeImgMetadata(node)
	for _, ci := range comment.images {
		if opts.Verbose {
			fmt.Println(sgfutils.BoardToString(*node.Board()))
		}

		cr, originalImgSize := calculateCrop([]*sgf.Node{node}, *opts)
		if opts.Verbose {
			fmt.Printf("crop to %v with original size %d\n", cr, originalImgSize)
		}

		fn := exportedImgFilename(sgfFilename, ci.name, "", string(opts.ImageType))
		switch opts.ImageType {
		case SVG:
			svg := draw2dsvg.NewSvg()
			if cr.isCrop() {
				band := float64(originalImgSize) / float64(boardSize)
				left := float64(cr.Left) * band
				up := float64(cr.Up) * band
				width := float64(boardSize-cr.Left-cr.Right)*band + 1
				height := float64(boardSize-cr.Up-cr.Down)*band + 1
				svg.ViewBox = fmt.Sprintf("%d %d %d %d", int(left), int(up), int(width), int(height)) // int(left), int(down), 100, 100) //int(right-left), int(down-up))
				if opts.Verbose {
					fmt.Printf("SVG crop %#v\n", cr)
					fmt.Printf("SVG viewbox %#v\n", svg.ViewBox)
				}
			} else {
				svg.Width = fmt.Sprint(originalImgSize)
				svg.Height = fmt.Sprint(originalImgSize)
			}

			boardToImage(draw2dsvg.NewGraphicContext(svg), *node, originalImgSize)
			byts, err := xml.Marshal(svg)
			if err != nil {
				return nil, err
			}
			files = append(files, GobanImageFile{Name: fn, Contents: byts})
		case PNG:
			dest := image.NewRGBA(image.Rect(0, 0, int(originalImgSize), int(originalImgSize)))
			gc := draw2dimg.NewGraphicContext(dest)
			boardToImage(gc, *node, originalImgSize)

			var i image.Image
			if opts.Grayscale {
				i = cropImage(grayscale(dest, originalImgSize), cr, *node.Board(), originalImgSize)
			} else {
				i = cropImage(dest, cr, *node.Board(), originalImgSize)
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

	animations, err := saveAnimations(comment, node, opts, sgfFilename, depth)
	if err != nil {
		return nil, err
	}
	files = append(files, animations...)

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

func grayscale(dest image.Image, originalImgSize int) image.Image {
	w, h := originalImgSize, originalImgSize
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

func saveAnimations(cm nodeImgMetdata, node *sgf.Node, opts *Options, sgfFilename string, depth int) ([]GobanImageFile, error) {
	var res []GobanImageFile
	for _, ca := range cm.animate {
		tmpNode := node
		var parentImage imgMetadata

		animatedNodes := []*sgf.Node{tmpNode}

	loop:
		for {
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
				return nil, fmt.Errorf("can't find node with img name '%s' for animation (loc %d)", ca.name, depth+1)
			} else {
				animatedNodes = append([]*sgf.Node{tmpNode}, animatedNodes...)
			}
		}

		cr, originalSize := calculateCrop(animatedNodes, *opts)
		_ = parentImage

		fn := exportedImgFilename(sgfFilename, ca.name, "animated", string(opts.ImageType))
		switch opts.ImageType {
		case PNG:
			var images []image.Image
			for _, n := range animatedNodes {
				img := image.NewRGBA(image.Rect(0, 0, originalSize, originalSize))
				gc := draw2dimg.NewGraphicContext(img)
				boardToImage(gc, *n, originalSize)
				images = append(images, img)
			}
			fmt.Printf("Found %d images to animate\n", len(images))
			for n := range images {
				if opts.Grayscale {
					images[n] = grayscale(images[n], originalSize)
				}
				images[n] = cropImage(images[n], cr, *node.Board(), originalSize)
			}
			byts, err := animatePng(images, fn)
			if err != nil {
				return nil, err
			}
			res = append(res, GobanImageFile{Name: fn, Contents: byts})
		}
		fmt.Printf("Saved %d board positions on move %d (%s) to: %s\n", len(animatedNodes), depth, ca.name, fn)
	}

	return res, nil
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

func cropImage(img image.Image, c Crop, board sgf.Board, originalImgSize int) image.Image {
	if !c.isCrop() {
		return img
	}
	band := float64(originalImgSize) / float64(board.Size)
	left := float64(c.Left) * band
	right := float64(c.Right) * band
	up := float64(c.Up) * band
	down := float64(c.Down) * band
	if left+right > float64(originalImgSize) {
		left = float64(originalImgSize) / 2.
		right = float64(originalImgSize) / 2.
	}
	if up+down > float64(originalImgSize) {
		up = float64(originalImgSize) / 2.
		down = float64(originalImgSize) / 2.
	}
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(image.Rect(int(left), int(up), int(originalImgSize)-int(right), int(originalImgSize)-int(down)))
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
