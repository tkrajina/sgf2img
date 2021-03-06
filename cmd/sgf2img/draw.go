package main

import (
	"bufio"
	"fmt"
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

func boardToImage(gc draw2d.GraphicContext, node sgf.Node, opts ctx) {
	gc.SetFillColor(color.RGBA{239, 193, 113, 0xff})
	gc.BeginPath()  // Initialize a new path
	gc.MoveTo(0, 0) // Move to a position to start the new path
	gc.LineTo(0, float64(opts.imageSize))
	gc.LineTo(float64(opts.imageSize), float64(opts.imageSize))
	gc.LineTo(float64(opts.imageSize), 0)
	gc.Close()
	gc.FillStroke()
	for i := 0; i < node.Board().Size; i++ {
		gc.SetStrokeColor(color.RGBA{0x00, 0x00, 0x00, 0xff})
		gc.SetLineWidth(0.20)

		gc.MoveTo(boardCoordinateToImageCoordinate(i, 0, int(opts.imageSize), *node.Board()))
		gc.LineTo(boardCoordinateToImageCoordinate(i, node.Board().Size-1, int(opts.imageSize), *node.Board()))
		gc.Close()
		gc.FillStroke()

		gc.MoveTo(boardCoordinateToImageCoordinate(0, i, int(opts.imageSize), *node.Board()))
		gc.LineTo(boardCoordinateToImageCoordinate(node.Board().Size-1, i, int(opts.imageSize), *node.Board()))
		gc.Close()
		gc.FillStroke()
	}

	band := float64(opts.imageSize) / float64(node.Board().Size)

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
		x, y := boardCoordinateToImageCoordinate(h[0], h[1], int(opts.imageSize), *node.Board())
		gc.SetFillColor(image.Black)
		gc.ArcTo(x, y, band/10, band/10, 0, 2*math.Pi)
		gc.Close()
		gc.FillStroke()
	}

	// Stones
	drawStones(gc, node, int(opts.imageSize))

	var lastMoves []string
	if whiteMove, ok := node.GetValue(sgfutils.SGFTagWhiteMove); ok {
		lastMoves = append(lastMoves, whiteMove)
	}
	if blackMove, ok := node.GetValue(sgfutils.SGFTagBlackMove); ok {
		lastMoves = append(lastMoves, blackMove)
	}
	for _, circle := range lastMoves {
		x, y := sgfCoordinatesToImageCoordinates(circle, int(opts.imageSize), *node.Board())
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
	drawPolyline(node.AllValues("TR"), gc, &node, int(opts.imageSize), 3, 0)
	// squares:
	drawPolyline(node.AllValues("SQ"), gc, &node, int(opts.imageSize), 4, math.Pi/4)

	drawLabels(gc, node, int(opts.imageSize))
	drawCircles(gc, node, int(opts.imageSize))
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

func nodeAnnotationsToAnki(node sgf.Node) string {
	var res []string
	tags := []string{
		/*SGFTagBlackMove, SGFTagWhiteMove, <-- not needed because they are represented with WB (uppercase) */
		sgfutils.SGFTagComment,
		sgfutils.SGFTagTriangle,
		sgfutils.SGFTagSquare,
		sgfutils.SGFTagCircle,
		sgfutils.SGFTagX,
		sgfutils.SGFTagLabel,
	}
	valuesPerTag := map[string][]string{}
	for _, tag := range tags {
		for _, val := range node.AllValues(tag) {
			valuesPerTag[tag] = append(valuesPerTag[tag], val)
		}
	}

	for _, tag := range tags {
		val := strings.Join(valuesPerTag[tag], ",")
		var cleaned []string
	lines_loop:
		for _, line := range strings.Split(val, "\n") {
			for _, d := range directives {
				if strings.HasPrefix(strings.TrimSpace(line), d) {
					continue lines_loop
				}
			}
			cleaned = append(cleaned, line)
		}
		val = strings.TrimSpace(strings.Join(cleaned, "\\n"))
		if val != "" {
			res = append(res, fmt.Sprintf("%s:%s", tag, val))
		}
	}
	return strings.Join(res, "\n") + "\n"
}

func boardAnnotationsToAnki(node sgf.Node) string {
	var res []string
	tags := []string{
		sgfutils.SGFTagWhiteName,
		sgfutils.SGFTagBlackName,
		sgfutils.SGFTagApplication,
		sgfutils.SGFTagDate,
		sgfutils.SGFTagBlackRank,
		sgfutils.SGFTagWhiteRank,
		sgfutils.SGFTagResult,
	}
	for _, tag := range tags {
		for _, val := range node.AllValues(tag) {
			if val != "" {
				res = append(res, fmt.Sprintf("%s:%s", tag, val))
			}
		}
	}
	return strings.Join(res, "\n") + "\n"
}

func boardToAnki(node sgf.Node) string {
	// ????????
	var res []string
	board := node.Board()
	for i := 0; i < board.Size; i++ {
		res = append(res, "")
		for j := 0; j < board.Size; j++ {
			color := board.Get(sgf.Point(j, i))
			latestMove := false
			for _, tag := range []string{sgfutils.SGFTagBlackMove, sgfutils.SGFTagWhiteMove} {
				if value, found := node.GetValue(tag); found && value != "" {
					x, y, _ := sgf.ParsePoint(value, node.Board().Size)
					if x == j && y == i {
						//fmt.Println(x, y, i, y)
						latestMove = true
					}
				}
			}

			pt := sgfutils.Empty
			switch color {
			case sgf.WHITE:
				pt = sgfutils.WhiteCircle
			case sgf.BLACK:
				pt = sgfutils.BlackCircle
			}

			if latestMove {
				pt = strings.ToUpper(pt)
			}
			res[len(res)-1] += pt
		}
	}
	return strings.Join(res, "\n")
}
