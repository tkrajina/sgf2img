package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

type stringFlags []string

func (sf *stringFlags) String() string {
	return "my string representation"
}

func (sf *stringFlags) Set(value string) error {
	*sf = append(*sf, value)
	return nil
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Var(&removePrefixes, "rm", "Remove prefix")
	flag.Var(&deleteLinesWithPrefixes, "rmln", "Remove lines with prefix")
	flag.Parse()

	for _, fn := range flag.Args() {
		s, err := sgf.Load(fn)
		panicIfErr(err)

		panicIfErr(CleanComments(s))

		ext := filepath.Ext(fn)
		fn := fn[0:len(fn)-len(ext)] + "_cleaned.sgf"
		panicIfErr(s.Save(fn))
		fmt.Println("Saved to", fn)
		fmt.Println(s.SGF())
	}
}

var (
	removePrefixes          stringFlags
	deleteLinesWithPrefixes stringFlags
)

func CleanComments(node *sgf.Node) error {

	comments := node.AllValues(sgfutils.SGFTagComment)
	for n, comment := range comments {
		var lines []string
	_lines_loop:
		for _, line := range strings.Split(comment, "\n") {
			for _, removePrefix := range removePrefixes {
				removePrefix = strings.TrimSpace(removePrefix)
				if len(removePrefix) > 0 && strings.HasPrefix(line, removePrefix) {
					line2 := line[len(removePrefix):]
					fmt.Println("replacing:", line, "->", line2)
					lines = append(lines, line2)
					continue _lines_loop
				}
			}
			for _, deleteLineWithPrefix := range deleteLinesWithPrefixes {
				deleteLineWithPrefix = strings.TrimSpace(deleteLineWithPrefix)
				if len(deleteLineWithPrefix) > 0 && strings.HasPrefix(line, deleteLineWithPrefix) {
					fmt.Println("deleting:", line)
					continue _lines_loop
				}
			}
			lines = append(lines, line)
		}
		newComment := strings.Join(lines, "\n")
		fmt.Println("Changing comment:")
		fmt.Println(comment)
		fmt.Println("...to:")
		fmt.Println(newComment)
		fmt.Println()
		comments[n] = newComment
	}
	node.SetValues(sgfutils.SGFTagComment, comments)

	for _, child := range node.Children() {
		if err := CleanComments(child); err != nil {
			return err
		}
	}

	return nil
}
