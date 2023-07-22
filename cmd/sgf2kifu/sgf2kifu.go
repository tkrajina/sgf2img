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
	page              int
	continuousCounter bool
}

func main() {
	if err := doStuff(); err != nil {
		panic(err)
	}
}

func doStuff() error {
	var opt opt
	flag.IntVar(&opt.page, "p", 20, "Nomber of moves per page")
	flag.BoolVar(&opt.continuousCounter, "c", false, "continuous counter (otherwise it starts from 1 on every page)")
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
	gi := sgfutils.ParseGameInfo(node)
	b := bytes.NewBuffer([]byte{})
	title := gi.BlackName + " vs " + gi.WhiteName
	_, _ = b.WriteString("# " + title + "\n\n")

	infoIfNonEmpty("Date:", gi.Date, b)
	infoIfNonEmpty("Event:", gi.Event, b)
	infoIfNonEmpty("BlackName:", gi.BlackName, b)
	infoIfNonEmpty("BlackRank:", gi.BlackRank, b)
	infoIfNonEmpty("BlackTeam:", gi.BlackTeam, b)
	infoIfNonEmpty("WhiteName:", gi.WhiteName, b)
	infoIfNonEmpty("WhiteRank:", gi.WhiteRank, b)
	infoIfNonEmpty("WhiteTeam:", gi.WhiteTeam, b)
	infoIfNonEmpty("Result:", gi.Result, b)
	infoIfNonEmpty("Rules:", gi.Rules, b)
	infoIfNonEmpty("Komi:", gi.Komi, b)
	infoIfNonEmpty("Handicap:", gi.Handicap, b)

	_, _ = b.WriteString("\n\n")
	tmpNode := node
	n := 0
	moves := map[string][]int{}
	for {
		wm, _ := tmpNode.GetValue(sgfutils.SGFTagWhiteMove)
		bm, _ := tmpNode.GetValue(sgfutils.SGFTagBlackMove)
		coord := ""
		if wm != "" {
			coord = wm
		} else if bm != "" {
			coord = bm
		}
		moves[coord] = append(moves[coord], n)

		if len(moves[coord]) > 1 {
			fmt.Printf("Move #%d %s has more than one moves: %#v\n", n+1, coord, moves[coord])
		}

		if n > 0 && n%opt.page == 0 {
			imgNotes = append(imgNotes, toImage(tmpNode, moves, n, opt))
			moves = make(map[string][]int)
		}

		if len(tmpNode.Children()) == 0 {
			if len(moves) > 0 {
				imgNotes = append(imgNotes, toImage(tmpNode, moves, n, opt))
			}
			break
		}
		tmpNode = tmpNode.Children()[0]
		n += 1
	}

	_, files, err := sgf2img.ProcessSGF(fn, node, &sgf2img.Options{ImageType: sgf2img.SVG, ImageSize: 1000, BW: true})
	if err != nil {
		return err
	}
	for n, file := range files {
		// if err := os.WriteFile(file.Name, file.Contents, 0700 /*  */); err != nil {
		// 	return err
		// }
		if len(imgNotes[n]) > 0 {
			_, _ = b.WriteString("## " + imgNotes[n][0] + "\n\n")
		}
		_, _ = b.WriteString(fmt.Sprintf(`<div style="width: %dpx;">`, 1000 /* TODO */))
		_, _ = b.Write(file.Contents)
		_, _ = b.WriteString("</div>\n\n")
		if len(imgNotes[n]) > 1 {
			_, _ = b.WriteString(strings.Join(imgNotes[n][1:], "\n"))
		}
		_, _ = b.WriteString("\n\n")
		_, _ = b.WriteString(`

<div style="page-break-after: always;"></div>

`)
	}
	infoIfNonEmpty("Result:", gi.Result, b)

	// os.WriteFile("kifu.md", b.Bytes(), 0700)

	output := markdown.ToHTML(b.Bytes(), nil, nil)
	output = append([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>`+title+`</title>
    <meta charset="utf-8">
	<meta name="viewport" content="initial-scale=1.0, width=device-width"/>
	<style>
	.circle {
		display: inline-block;
		width: 2em;
		height: 2em;
		font-size: 1em;
		border-radius: 50%;
		background-color: #ddd; /* You can change the background color here */
		border: 1px solid black;
		text-align: center;
		line-height: 30px;
		font-weight: bold;
	  }
	</style>
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

func infoIfNonEmpty(desc, value string, res *bytes.Buffer) {
	if value != "" {
		_, _ = res.WriteString(fmt.Sprintf(" * %s: **%s**\n", strings.Trim(desc, ":"), strings.TrimSpace(value)))
	}
}

func toImage(node *sgf.Node, moves map[string][]int, imgNo int, o opt) (notes []string) {
	min, max := -1, -1
	vals := []string{}
	for _, numbers := range moves {
		for _, n := range numbers {
			if min == -1 || n < min {
				min = n
			}
			if max == -1 || n > max {
				max = n
			}
		}
	}
	if min < 1 {
		min = 1
	}
	substract := min - 1
	if o.continuousCounter {
		substract = 0
	}
	for coord, numbers := range moves {
		if len(numbers) > 0 {
			number := numbers[0]
			vals = append(vals, coord+":"+fmt.Sprint(number-substract))
		}
		if len(numbers) > 1 {
			fmt.Printf("Multiple moves %#v\n", numbers)
			for n := 1; n < len(numbers); n++ {
				notes = append(notes, fmt.Sprintf("<!-- %04d (used for sorting) --> <span class=\"circle\">%d</span> AT <span class=\"circle\">%d</span><br/>", numbers[n], numbers[n]-substract, numbers[0]-substract))
			}
		}
	}
	sort.Strings(notes)
	notes = append([]string{fmt.Sprintf("Moves %d - %d", min, max)}, notes...)
	node.SetValues(sgfutils.SGFTagLabel, vals)
	node.SetValue(sgf2img.DirectiveImg, fmt.Sprintf("_img_%d", imgNo))
	return
}
