package sgf2img

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

var (
	directiveImg   = "img"
	directiveStart = "start"
	directiveEnd   = "end"
)

var r = regexp.MustCompile(`^(\d*)(\w)$`)

type cropInfo struct {
	left, right, up, down int
}

func (c cropInfo) isCrop() bool {
	return c.left >= 0 || c.right >= 0 || c.up >= 0 || c.down >= 0
}

type imgMetadata struct {
	name string
	crop cropInfo
}
type animationMetadata struct {
	name string
}

type nodeImgMetdata struct {
	comment string
	images  []imgMetadata
	start   []imgMetadata
	animate []animationMetadata
}

func parseNodeImgMetadata(node *sgf.Node) (cm nodeImgMetdata) {
	comment, found := node.GetValue(sgfutils.SGFTagComment)
	cm.comment = comment
	if found {
		parseComment(comment, node.Board().Size, &cm)
	}

	for _, val := range node.AllValues(directiveEnd) {
		cm.animate = append(cm.animate, animationMetadata{name: val})
	}
	for _, val := range node.AllValues(directiveStart) {
		cm.start = append(cm.start, imgMetadata{name: val})
	}
	for _, val := range node.AllValues(directiveImg) {
		cm.images = append(cm.images, imgMetadata{name: val})
	}

	return
}

func parseComment(comment string, boardSize int, cm *nodeImgMetdata) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) == 0 {
			continue
		}
		isImg := parts[0] == "!"+directiveImg
		isStart := parts[0] == "!"+directiveStart
		if isImg || isStart {
			ci := imgMetadata{}
			if len(parts) > 1 {
				ci.name = parts[1]
			}
			if len(parts) > 2 {
				for i := 2; i < len(parts); i++ {
					matches := r.FindAllStringSubmatch(parts[i], -1)
					if len(matches) == 1 {
						n, err := strconv.ParseInt(matches[0][1], 10, 32)
						if err != nil {
							n = 1 + int64(boardSize)/2
						}
						n = int64(boardSize) - n
						letter := matches[0][2]
						switch letter {
						case "u":
							ci.crop.down = int(n)
						case "d":
							ci.crop.up = int(n)
						case "l":
							ci.crop.right = int(n)
						case "r":
							ci.crop.left = int(n)
						default:
							fmt.Println("Uknown:", parts[i])
						}
					}
				}
			}
			if isImg {
				cm.images = append(cm.images, ci)
			}
			if isStart {
				cm.start = append(cm.start, ci)
			}
		}
		if parts[0] == "!"+directiveEnd {
			ca := animationMetadata{}
			if len(parts) > 1 {
				ca.name = parts[1]
			}
			cm.animate = append(cm.animate, ca)
		}
	}
}
