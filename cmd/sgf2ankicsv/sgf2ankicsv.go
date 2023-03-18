package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	mainFirst     bool
	leaveTmpFiles bool
)

func main() {
	flag.BoolVar(&mainFirst, "m", false, "Main branch first?")
	flag.BoolVar(&leaveTmpFiles, "l", false, "Leave temp files?")
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	var csvRows [][]string

	for fileNo, fn := range flag.Args() {
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
			cleanedComment := []string{}
			ankiLines := []string{}
			for _, line := range strings.Split(comment, "\n") {
				if strings.HasPrefix(strings.ToLower(comment), "!anki") {
					ankiLines = append(ankiLines, line)
				} else {
					cleanedComment = append(cleanedComment, line)
				}
			}
			if variantNo == 0 {
				ankiLines = append(ankiLines, "!anki")
			}
			for ankiLineNo, ankiLine := range ankiLines {
				fmt.Printf("variant %d\n", variantNo)
				root.SetValue(sgfutils.SGFTagSource, fmt.Sprintf("%s variant %d", fn, variantNo))
				leafNode.MakeMainLine()
				leafNode.SetValue(sgfutils.SGFTagComment, strings.Join(cleanedComment, "\n")+"\n"+ankiLine)

				tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("sgf2anki_%d_%d_%d.sgf", fileNo, variantNo, ankiLineNo))
				if err != nil {
					return err
				}
				if err := root.Save(tmpFileName); err != nil {
					return err
				}

				byts, err := ioutil.ReadFile(tmpFileName)
				if err != nil {
					return err
				}

				if leaveTmpFiles {
					fmt.Println("Variant saved to ", tmpFileName)
				} else {
					_ = os.Remove(tmpFileName)
				}

				csvRows = append(csvRows, []string{string(byts)})
			}

			leafNode.SetValue(sgfutils.SGFTagComment, comment)
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
