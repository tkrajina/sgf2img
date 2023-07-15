package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
	"github.com/tkrajina/sgf2img/sgfutils/sgf2img"
)

type opt struct {
	page int
}

func main() {
	if err := doStuff(); err != nil {
		panic(err)
	}
}

func doStuff() error {
	var opt opt
	flag.IntVar(&opt.page, "p", 10, "Nomber of moves per page")
	flag.Parse()
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			return err
		}
		if err := nodeToKifu(fn, opt, node); err != nil {
			return err
		}
	}
	return nil
}

func nodeToKifu(fn string, opt opt, node *sgf.Node) (err error) {
	var imgNotes [][]string
	gameInfo := sgfutils.ParseGameInfo(node)
	b := bytes.NewBuffer([]byte{})
	title := gameInfo.BlackName + " vs " + gameInfo.WhiteName + " (" + gameInfo.Result + ")"
	_, _ = b.WriteString("# " + title + "\n\n")
	tmpNode := node
	n := 0
	moves := map[string][]int{}
	for {
		println(n)
		wm, _ := tmpNode.GetValue(sgfutils.SGFTagWhiteMove)
		bm, _ := tmpNode.GetValue(sgfutils.SGFTagBlackMove)
		if wm != "" {
			moves[wm] = append(moves[wm], n)
		} else if bm != "" {
			moves[bm] = append(moves[bm], n)
		}

		if n > 0 && n%opt.page == 0 {
			imgNotes = append(imgNotes, toImage(tmpNode, moves, n))
			moves = make(map[string][]int)
		}

		if len(tmpNode.Children()) == 0 {
			if len(moves) > 0 {
				imgNotes = append(imgNotes, toImage(tmpNode, moves, n))
			}
			break
		}
		tmpNode = tmpNode.Children()[0]
		n += 1
	}

	fmt.Printf("%#v\n", moves)

	_, files, err := sgf2img.ProcessSGF(fn, node, &sgf2img.Options{ImageType: sgf2img.SVG, ImageSize: 1000})
	if err != nil {
		return err
	}
	for n, file := range files {
		if err := os.WriteFile(file.Name, file.Contents, 0700 /*  */); err != nil {
			return err
		}
		_, _ = b.WriteString(fmt.Sprintf(`<div style="width: %dpx;">`, 1000 /* TODO */))
		_, _ = b.Write(file.Contents)
		_, _ = b.WriteString("</div>\n\n")
		_, _ = b.WriteString(strings.Join(imgNotes[n], "\n"))
		_, _ = b.WriteString("\n\n")
	}

	// os.WriteFile("kifu.md", b.Bytes(), 0700)

	output := markdown.ToHTML(b.Bytes(), nil, nil)
	output = append([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>`+title+`</title>
    <meta charset="utf-8">
	<meta name="viewport" content="initial-scale=1.0, width=device-width"/>
</head>
<body>
		`), output...)
	output = append(output, []byte("\n</html>")...)

	ext := filepath.Ext(fn)
	outputFn := fn[0:len(fn)-len(ext)] + "_kifu.html"
	if err := os.WriteFile(outputFn, output, 0700); err != nil {
		return err
	}
	fmt.Println("Written to ", outputFn)

	return nil
}

func toImage(node *sgf.Node, moves map[string][]int, imgNo int) (notes []string) {
	vals := []string{}
	for coord, numbers := range moves {
		if len(numbers) > 0 {
			vals = append(vals, coord+":"+fmt.Sprint(numbers[0]))
		}
		if len(numbers) > 1 {
			for n := 0; n < len(numbers)-1; n++ {
				notes = append(notes, fmt.Sprintf("* <!-- %04d (used for sorting) --> (%d) at (%d)", numbers[len(numbers)-1], numbers[len(numbers)-1], numbers[n]))
			}
		}
	}
	sort.Strings(notes)
	node.SetValues(sgfutils.SGFTagLabel, vals)
	node.SetValue(sgf2img.DirectiveImg, fmt.Sprintf("_img_%d", imgNo))
	return
}
