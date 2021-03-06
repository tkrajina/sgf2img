package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

var (
	directiveImg     = []string{"!s", "!start", "!img"}
	directiveAnimate = []string{"!e", "!end", "!animate"}
)

var directives = []string{}

func init() {
	directives = append(directives, directiveImg...)
	directives = append(directives, directiveAnimate...)
}

var r = regexp.MustCompile(`^(\d*)(\w)$`)

type commentImage struct {
	name                  string
	left, right, up, down int
}
type commentAnimate struct {
	name string
}

type commentMedatada struct {
	comment string
	images  []commentImage
	animate []commentAnimate
}

func parseNodeComment(node *sgf.Node) (cm commentMedatada) {
	comment, found := node.GetValue(sgfutils.SGFTagComment)
	cm.comment = comment
	if found {
		cm = parseComment(comment, node.Board().Size)
	}
	return
}

func parseComment(comment string, boardSize int) (cm commentMedatada) {
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		line = strings.ToLower(line)
		parts := strings.Fields(strings.TrimSpace(line))
		if len(parts) == 0 {
			continue
		}
		if isOneOf(parts[0], directiveImg) {
			ci := commentImage{}
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
							ci.down = int(n)
						case "d":
							ci.up = int(n)
						case "l":
							ci.right = int(n)
						case "r":
							ci.left = int(n)
						default:
							fmt.Println("Uknown:", parts[i])
						}
					}
				}
			}
			cm.images = append(cm.images, ci)
		}
		if isOneOf(parts[0], directiveAnimate) {
			ca := commentAnimate{}
			if len(parts) > 1 {
				ca.name = parts[1]
			}
			cm.animate = append(cm.animate, ca)
		}
	}
	return
}
