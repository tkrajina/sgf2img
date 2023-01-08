package sgf2img

import (
	"bufio"
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
	"strings"

	"github.com/llgcode/draw2d"
	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func boardToImage(gc draw2d.GraphicContext, node sgf.Node, imgSize int) {
	gc.SetFillColor(color.RGBA{239, 193, 113, 0xff})
	gc.BeginPath()  // Initialize a new path
	gc.MoveTo(0, 0) // Move to a position to start the new path
	gc.LineTo(0, float64(imgSize))
	gc.LineTo(float64(imgSize), float64(imgSize))
	gc.LineTo(float64(imgSize), 0)
	gc.Close()
	gc.FillStroke()
	for i := 0; i < node.Board().Size; i++ {
		gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		gc.SetLineWidth(0.20)

		gc.MoveTo(boardCoordinateToImageCoordinate(i, 0, int(imgSize), *node.Board()))
		gc.LineTo(boardCoordinateToImageCoordinate(i, node.Board().Size-1, int(imgSize), *node.Board()))
		gc.Close()
		gc.FillStroke()

		gc.MoveTo(boardCoordinateToImageCoordinate(0, i, int(imgSize), *node.Board()))
		gc.LineTo(boardCoordinateToImageCoordinate(node.Board().Size-1, i, int(imgSize), *node.Board()))
		gc.Close()
		gc.FillStroke()
	}

	band := float64(imgSize) / float64(node.Board().Size)

	var hoshi [][2]int
	switch node.Board().Size {
	case 9:
		hoshi = [][2]int{{2, 2}}
	case 13:
		hoshi = [][2]int{{3, 3}}
	case 19:
		hoshi = [][2]int{
			{3, 3}, {3, 9}, {3, 15},
			{9, 3}, {9, 9}, {9, 15},
			{15, 3}, {15, 9}, {15, 15},
		}
	}
	for _, h := range hoshi {
		x, y := boardCoordinateToImageCoordinate(h[0], h[1], int(imgSize), *node.Board())
		gc.SetFillColor(image.Black)
		gc.ArcTo(x, y, band/10, band/10, 0, 2*math.Pi)
		gc.Close()
		gc.FillStroke()
	}

	// Stones
	drawStones(gc, node, int(imgSize))

	var lastMoves []string
	if whiteMove, ok := node.GetValue(sgfutils.SGFTagWhiteMove); ok {
		lastMoves = append(lastMoves, whiteMove)
	}
	if blackMove, ok := node.GetValue(sgfutils.SGFTagBlackMove); ok {
		lastMoves = append(lastMoves, blackMove)
	}
	for _, circle := range lastMoves {
		x, y := sgfCoordinatesToImageCoordinates(circle, int(imgSize), *node.Board())
		if node.Board().Get(circle) == sgf.BLACK {
			gc.SetFillColor(color.White)
			gc.SetStrokeColor(color.White)
		} else {
			gc.SetFillColor(color.Black)
			gc.SetStrokeColor(color.Black)
		}
		gc.SetLineWidth(0.40)
		gc.ArcTo(x, y, band/8, band/8, 0, 2*math.Pi)
		gc.Close()
		gc.FillStroke()
	}

	// triangles:
	drawPolyline(node.AllValues("TR"), gc, &node, int(imgSize), 3, 0)
	// squares:
	drawPolyline(node.AllValues("SQ"), gc, &node, int(imgSize), 4, math.Pi/4)

	drawLabels(gc, node, int(imgSize))
	drawCircles(gc, node, int(imgSize))
}

func drawStones(gc draw2d.GraphicContext, node sgf.Node, imgSize int) {
	band := float64(imgSize) / float64(node.Board().Size) * .9
	for i := 0; i < node.Board().Size; i++ {
		for j := 0; j < node.Board().Size; j++ {
			gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
			gc.SetLineWidth(0.5)
			color := node.Board().Get(sgf.Point(i, j))
			var fillColor *image.Uniform
			switch color {
			case sgf.EMPTY:
			case sgf.WHITE:
				fillColor = image.White
			case sgf.BLACK:
				fillColor = image.Black
			}
			if fillColor != nil {
				x, y := boardCoordinateToImageCoordinate(i, j, imgSize, *node.Board())
				gc.SetFillColor(fillColor)
				gc.ArcTo(x, y, band/2, band/2, 0, 2*math.Pi)
				gc.Close()
				gc.FillStroke()
			}
		}
	}
}

func drawCircles(gc draw2d.GraphicContext, node sgf.Node, imgSize int) {
	band := float64(imgSize) / float64(node.Board().Size) * .9
	for _, circle := range node.AllValues("CR") {
		x, y := sgfCoordinatesToImageCoordinates(circle, imgSize, *node.Board())
		if node.Board().Get(circle) == sgf.BLACK {
			gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0x00})
			gc.SetStrokeColor(color.White)
		} else {
			gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0x00})
			gc.SetStrokeColor(color.Black)
		}
		gc.SetLineWidth(0.40)
		gc.ArcTo(x, y, band/3, band/3, 0, 2*math.Pi)
		gc.Close()
		gc.FillStroke()
	}
}

func drawLabels(gc draw2d.GraphicContext, node sgf.Node, imgSize int) {
	band := float64(imgSize) / float64(node.Board().Size) * .9
	for _, label := range node.AllValues("LB") {
		parts := strings.Split(label, ":")
		if len(parts) != 2 {
			continue
		}
		x, y := sgfCoordinatesToImageCoordinates(parts[0], imgSize, *node.Board())
		txt := parts[1]

		if node.Board().Get(parts[0]) == sgf.BLACK {
			gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
		} else {
			gc.SetFillColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		}

		// Set the font luximbi.ttf
		gc.SetFontData(draw2d.FontData{Name: "goregular", Family: draw2d.FontFamilyMono})
		// Set the fill text color to black
		fontSize := band * .6
		gc.SetFontSize(fontSize)
		// Display Hello World
		gc.FillStringAt(txt, x-fontSize/2.5, y+fontSize/2)
	}
}

func drawPolyline(triangles []string, gc draw2d.GraphicContext, node *sgf.Node, imgSize, polylineSide int, initialAngle float64) {
	band := float64(imgSize) / float64(node.Board().Size) * .9
	for _, triangle := range triangles {
		if node.Board().Get(triangle) == sgf.BLACK {
			gc.SetStrokeColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
		} else {
			gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		}
		gc.SetLineWidth(0.2)
		boardX, boardY, onboard := sgf.ParsePoint(triangle, node.Board().Size)
		if !onboard {
			continue
		}
		x, y := boardCoordinateToImageCoordinate(boardX, boardY, imgSize, *node.Board())

		var pts [][2]float64
		for i := 0.0; i < 2*math.Pi; i += 2 * math.Pi / float64(polylineSide) {
			pts = append(pts, [2]float64{x - band/2.*math.Sin(i+initialAngle), y - band/2.*math.Cos(i+initialAngle)})
		}
		for n := range pts {
			gc.BeginPath()
			gc.MoveTo(pts[n][0], pts[n][1])
			gc.LineTo(pts[(n+1)%len(pts)][0], pts[(n+1)%len(pts)][1])
			gc.Close()
			gc.FillStroke()
		}
	}
}

func SaveToGifFile(filePath string, m image.Image) error {
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	// Create Writer from file
	b := bufio.NewWriter(f)
	// Write the image into the buffer
	err = gif.Encode(b, m, &gif.Options{})
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}
