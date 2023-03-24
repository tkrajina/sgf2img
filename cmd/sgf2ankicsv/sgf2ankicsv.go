package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

var (
	mainFirst            bool
	chunks               bool
	leaveTmpFiles        bool
	cleanKatrainComments bool
)

func main() {
	flag.BoolVar(&mainFirst, "m", false, "Main branch first?")
	flag.BoolVar(&leaveTmpFiles, "l", false, "Leave temp files?")
	flag.BoolVar(&chunks, "c", false, "Extract chunks (instead of leaving full sgf)?")
	flag.BoolVar(&cleanKatrainComments, "ck", false, "Clean KaTrain comments?")
	flag.Parse()
	panicIfErr(doStuff())
}

func parseComment(leafNode *sgf.Node) (string, []string, []string) {
	comments := leafNode.AllValues(sgfutils.SGFTagComment)
	cleanedComment := []string{}
	ankiLines := []string{}
	for _, comment := range comments {
		for _, line := range strings.Split(comment, "\n") {
			if strings.HasPrefix(strings.ToLower(line), "!anki") {
				ankiLines = append(ankiLines, line)
			} else {
				cleanedComment = append(cleanedComment, line)
			}
		}
	}
	return strings.Join(comments, "\n"), cleanedComment, ankiLines
}

func doStuff() error {
	var csvRows [][]string

	for fileNo, fn := range flag.Args() {
		root, err := sgf.Load(fn)
		if err != nil {
			fmt.Println("Error reading:", fn, "", err.Error())
			return err
		}

		if cleanKatrainComments {
			if err := sgfutils.CleanKatrainStuff(root); err != nil {
				return err
			}
		}

		ankiNodes := findAnkiNodes(root, 0)
		if err != nil {
			return err
		}
		fmt.Printf("Found %d variants\n", len(ankiNodes))

		for variantNo, leafNode := range ankiNodes {
			comment, cleanedComment, ankiLines := parseComment(leafNode)
			if !chunks && variantNo == 0 {
				ankiLines = append(ankiLines, "!anki")
			}
			for ankiLineNo, ankiLine := range ankiLines {
				fmt.Printf("anki line: %s\n", ankiLine)
				fmt.Printf("variant %d\n", variantNo)

				tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("sgf2anki_%d_%d_%d.sgf", fileNo, variantNo, ankiLineNo))
				if chunks {
					ankiLineParts := append(strings.Fields(ankiLine), "1000000")
					fmt.Printf("parts %#v\n", ankiLineParts)
					n, err := strconv.ParseInt(ankiLineParts[1], 10, 32)
					_ = n
					if err != nil {
						return err
					}
					fmt.Println("TODO", ankiLine)
					sub, err := sgfutils.Sub(leafNode, -int(n))
					if err != nil {
						return err
					}
					if err := sub.Save(tmpFileName); err != nil {
						return err
					}
				} else {
					root.SetValue(sgfutils.SGFTagSource, fmt.Sprintf("%s variant %d", fn, variantNo))
					leafNode.MakeMainLine()
					leafNode.SetValue(sgfutils.SGFTagComment, strings.Join(cleanedComment, "\n")+"\n"+ankiLine)
					if err != nil {
						return err
					}
					if err := root.Save(tmpFileName); err != nil {
						return err
					}
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

func findAnkiNodes(node *sgf.Node, depth int) (ankiNode []*sgf.Node) {
	_, _, ankiLines := parseComment(node)
	if len(node.Children()) == 0 {
		fmt.Printf("Leaf node at pos %d\n", depth)
		ankiNode = append(ankiNode, node)
	} else if len(ankiLines) > 0 {
		fmt.Printf("Anki node at pos %d\n", depth)
		ankiNode = append(ankiNode, node)
	}
	for _, child := range node.Children() {
		ankiNode = append(ankiNode, findAnkiNodes(child, depth+1)...)
	}
	return
}
