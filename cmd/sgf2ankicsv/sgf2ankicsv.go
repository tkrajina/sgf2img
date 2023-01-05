package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var subproblems bool

func main() {
	flag.BoolVar(&subproblems, "subproblems", false, "Extract subproblems")
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	var csvRows [][]string
sgf_loop:
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			fmt.Println("Error reading:", fn, "", err.Error())
			return err
		}

		if len(node.Children()) == 0 {
			fmt.Println("Empty sgf:", fn)
			continue sgf_loop
		}

		if subproblems {
			findSub(node)
		} else {
			node.SetValue("CROP", "auto")
			node.SetValue(sgfutils.SGFTagSource, fn)

			tmpFile, err := os.CreateTemp(os.TempDir(), "sgf2anki_*.sgf")
			if err != nil {
				return err
			}
			if err := node.Save(tmpFile.Name()); err != nil {
				return err
			}

			byts, err := ioutil.ReadFile(tmpFile.Name())
			if err != nil {
				return err
			}
			_ = os.Remove(tmpFile.Name())

			csvRows = append(csvRows, []string{string(byts)})
		}
	}

	fn := "anki.csv"
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	csvwriter := csv.NewWriter(f)
	if err := csvwriter.WriteAll(csvRows); err != nil {
		return err
	}
	csvwriter.Flush()

	fmt.Printf("Saved %d rows to %s\n", len(csvRows), fn)

	return nil
}

func findSub(n *sgf.Node) error {
	ch := n.Children()
	if len(ch) == 0 {
		return nil
	}
	for _, child := range ch {
		comment, found := child.GetValue(sgfutils.SGFTagComment)
		if found && len(comment) > 0 {
			for _, line := range strings.Split(comment, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "!{") && strings.HasSuffix(line, "}") {
					startNo, err := strconv.ParseInt(strings.Trim(line, "!{} \r\t\n"), 10, 32)
					if err != nil {
						return err
					}
					getSub(child, int(startNo))
				}
			}
		}
	}
	return nil
}

func getSub(n *sgf.Node, startFrom int) error {
	branch := []*sgf.Node{n}
	for n.Parent() != nil {
		n = n.Parent()
		branch = append([]*sgf.Node{n}, branch...)
	}

	branch = branch[startFrom:]
	for n := 0; n < len(branch); n++ {
		if n == 0 {
			branch[0] = branch[0].Copy()
		}
	}

	if len(branch) >= startFrom {
		return fmt.Errorf("invalid start %s", startFrom)
	}
}
