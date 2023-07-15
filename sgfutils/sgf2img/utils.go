package sgf2img

import (
	"math"
	"strings"

	"github.com/rooklift/sgf"
)

func expandCoordinates(coords []string, boardSize int) ([]string, error) {
	res := []string{}
	for _, coord := range coords {
		expanded, err := expandCoordinatesRange(coord, boardSize)
		if err != nil {
			return nil, err
		}
		res = append(res, expanded...)
	}
	return res, nil
}

func expandCoordinatesRange(coord string, boardSize int) ([]string, error) {
	parts := strings.Split(coord, ":")
	if len(parts) != 2 {
		return []string{coord}, nil
	}
	x1, y1, _ := sgf.ParsePoint(parts[0], boardSize)
	x2, y2, _ := sgf.ParsePoint(parts[1], boardSize)
	minX := int(math.Min(float64(x1), float64(x2)))
	maxX := int(math.Max(float64(x1), float64(x2)))
	minY := int(math.Min(float64(y1), float64(y2)))
	maxY := int(math.Max(float64(y1), float64(y2)))
	res := []string{}
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			res = append(res, sgf.Point(x, y))
		}
	}
	return res, nil
}

func isOneOf(s string, strs []string) bool {
	for _, str := range strs {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}
