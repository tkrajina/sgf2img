package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var mainFirst bool

func main() {
	flag.BoolVar(&mainFirst, "m", false, "Main branch first?")
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	var csvRows [][]string

	for _, fn := range flag.Args() {
		root, err := sgf.Load(fn)
		if err != nil {
			fmt.Println("Error reading:", fn, "", err.Error())
			return err
		}

		leafNodes := findLeafNodes(root, 0)
		if err != nil {
			return err
		}
		fmt.Printf("Found %d variants\n", len(leafNodes))

		for variantNo, leafNode := range leafNodes {
			comment, _ := leafNode.GetValue(sgfutils.SGFTagComment)
			if variantNo == 0 || strings.Contains(strings.ToLower(comment), "!anki") {
				fmt.Printf("variant %d\n", variantNo)
				root.SetValue(sgfutils.SGFTagSource, fmt.Sprintf("%s variant %d", fn, variantNo))
				leafNode.MakeMainLine()

				tmpFile, err := os.CreateTemp(os.TempDir(), "sgf2anki_*.sgf")
				if err != nil {
					return err
				}
				if err := root.Save(tmpFile.Name()); err != nil {
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
	}

	if !mainFirst {
		for n := 0; n < len(csvRows)/2; n++ {
			csvRows[n], csvRows[len(csvRows)-n-1] = csvRows[len(csvRows)-n-1], csvRows[n]
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

func findLeafNodes(node *sgf.Node, depth int) (leafNodes []*sgf.Node) {
	if len(node.Children()) == 0 {
		fmt.Printf("Leaf node at pos %d\n", depth)
		leafNodes = append(leafNodes, node)
	} else {
		for _, child := range node.Children() {
			leafNodes = append(leafNodes, findLeafNodes(child, depth+1)...)
		}
	}
	return leafNodes
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

	return nil
}
