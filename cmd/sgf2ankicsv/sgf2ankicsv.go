package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type intFlags []int

func (sf *intFlags) String() string {
	return fmt.Sprintf("%v", *sf)
}

func (sf *intFlags) Set(value string) error {
	for _, val := range strings.Split(value, ",") {
		n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 32)
		if err != nil {
			fmt.Println("Invalid number(s):", value)
		}
		*sf = append(*sf, int(n))
	}
	return nil
}

var (
	mainFirst                      bool
	chunks                         bool
	leaveTmpFiles                  bool
	cleanKatrainComments           bool
	outFile                        string
	appendToFile                   string
	openWith                       string
	branchesAreSolutionsToMistakes bool
	branchesNo                     intFlags
	firstToPlay                    string
	tagsList                       string
)

// Mistakes line
// Mark mistake with X
// Edit and save edited
// Edit and if empty -- ignore
// Only specific color

func main() {
	flag.BoolVar(&mainFirst, "mf", false, "Main branch first?")
	flag.BoolVar(&leaveTmpFiles, "l", false, "Leave temp files?")
	flag.StringVar(&openWith, "e", "", "Open/edit with before adding?")
	flag.BoolVar(&chunks, "c", false, "Extract chunks (instead of leaving full sgf)?")
	flag.BoolVar(&cleanKatrainComments, "ck", false, "Clean KaTrain comments?")
	flag.StringVar(&outFile, "o", "", "Output file")
	flag.StringVar(&appendToFile, "a", "", "Append to existing file")
	flag.BoolVar(&branchesAreSolutionsToMistakes, "b", false, "Branches are solutions to problems")
	flag.Var(&branchesNo, "bn", "Branches at positions are solutions to problems (comma separated numbers)")
	flag.StringVar(&firstToPlay, "p", "", "First to play (color: w, b or player name)")
	flag.StringVar(&tagsList, "t", "", "Tags (coma separated)")
	flag.Parse()
	panicIfErr(doStuff())
}

func isMistake(node *sgf.Node) bool {
	comments := node.AllValues(sgfutils.SGFTagComment)
	for _, comment := range comments {
		for _, line := range strings.Split(comment, "\n") {
			if strings.HasPrefix(strings.ToLower(line), "!mistake") {
				if len(node.Parent().Children()) > 1 {
					return true
				}
				fmt.Println("Mistake without solution branch!")
				return false
			}
		}
	}
	if branchesAreSolutionsToMistakes && node.Parent() != nil && len(node.Parent().Children()) > 1 {
		return true
	}
	return false
}

func mistakesToAnki(node *sgf.Node) error {
	tmpNode := node
	for len(tmpNode.Children()) > 0 {
		parentNode := tmpNode.Parent()
		if isMistake(tmpNode) {
			branch := parentNode.Children()[1]
			if coord, ok := tmpNode.GetValue(sgfutils.SGFTagWhiteMove); ok {
				parentNode.AddValue(sgfutils.SGFTagX, coord)
			}
			if coord, ok := tmpNode.GetValue(sgfutils.SGFTagBlackMove); ok {
				parentNode.AddValue(sgfutils.SGFTagX, coord)
			}
			branchNode := branch
			branchLength := 1
			for len(branchNode.Children()) > 0 {
				branchLength++
				branchNode = branchNode.Children()[0]
			}
			branchNode.AddValue(sgfutils.SGFTagComment, fmt.Sprintf("!anki -%d", branchLength))
		}
		tmpNode = tmpNode.Children()[0]
	}
	return nil
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
	var sgfs []string

	dir := path.Join(os.TempDir(), fmt.Sprintf("sgf_%d", time.Now().Unix()))
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if leaveTmpFiles {
		fmt.Println("Saving files to:", dir)
	}

	for _, fn := range flag.Args() {
		fmt.Println("Loading", fn)
		root, err := sgf.Load(fn)
		if err != nil {
			fmt.Println("Error reading:", fn, "", err.Error())
			return err
		}

		if err := mistakesToAnki(root); err != nil {
			return err
		}

		if cleanKatrainComments {
			if err := sgfutils.CleanKatrainStuff(root); err != nil {
				return err
			}
		}

		var tags []string
		parts := strings.Split(path.Clean(fn), string(os.PathSeparator))
		for n := range parts {
			if n == 0 {
				continue
			}
			tagFn := strings.Join(parts[:n], string(os.PathSeparator)) + string(os.PathSeparator) + "anki_tag"
			byts, err := os.ReadFile(tagFn)
			if err != nil {
				continue
			}
			tag := strings.TrimSpace(string(byts))
			fmt.Println("found tag", tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		for _, tag := range strings.Split(tagsList, ",") {
			tag = strings.ToLower(strings.TrimSpace(tag))
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		root.SetValues("TAGS", tags)

		ankiNodes := findAnkiNodes(root, 0)
		if err != nil {
			return err
		}
		fmt.Printf("Found %d variants\n", len(ankiNodes))

		for variantNo, leafNode := range ankiNodes {
			root.SetValue(sgfutils.SGFTagSource, fmt.Sprintf("%s variant %d", fn, variantNo))
			comment, cleanedComment, ankiLines := parseComment(leafNode)
			if !chunks && variantNo == 0 {
				ankiLines = append(ankiLines, "!anki")
			}
			var resultSgf string
			for _, ankiLine := range ankiLines {
				fmt.Printf("anki line: %s\n", ankiLine)
				fmt.Printf("variant %d\n", variantNo)

				if chunks {
					ankiLineParts := append(strings.Fields(ankiLine), "1000000")
					fmt.Printf("parts %#v\n", ankiLineParts)
					n, err := strconv.ParseInt(ankiLineParts[1], 10, 32)
					_ = n
					if err != nil {
						return err
					}
					fmt.Println("TODO", ankiLine)
					sub, err := sgfutils.Sub(leafNode, -int(n), sgfutils.SubOpts{})
					if err != nil {
						return err
					}
					resultSgf = sub.SGF()
				} else {
					leafNode.MakeMainLine()
					leafNode.SetValue(sgfutils.SGFTagComment, strings.Join(cleanedComment, "\n")+"\n"+ankiLine)
					if err != nil {
						return err
					}
					resultSgf = root.SGF()
				}

				tmpNode, err := sgf.LoadSGF(resultSgf)
				if err != nil {
					return err
				}
				color, name := sgfutils.FindFirstMove(tmpNode)
				if firstToPlay != "" {
					if strings.EqualFold(firstToPlay, color) {
						fmt.Println("color", color, "OK")
					} else if strings.HasPrefix(strings.ToLower(firstToPlay), strings.ToLower(name)) {
						fmt.Println("player", name, "OK")
					} else {
						fmt.Println("first to play not ok => ignored")
						continue
					}
				}
				sgfs = append(sgfs, resultSgf)
			}

			leafNode.SetValue(sgfutils.SGFTagComment, comment)
		}
	}

	if !mainFirst {
		for n := 0; n < len(sgfs)/2; n++ {
			sgfs[n], sgfs[len(sgfs)-n-1] = sgfs[len(sgfs)-n-1], sgfs[n]
		}
	}

	// Write collection:
	tmpFileName := path.Join(dir, fmt.Sprintf("sgf2anki_%d.sgf", time.Now().UnixMilli()))
	if err := os.WriteFile(tmpFileName, []byte(strings.Join(sgfs, "\n")), 0700); err != nil {
		return err
	}
	if leaveTmpFiles {
		fmt.Println("Collection saved to ", tmpFileName)
	} else {
		defer func() { _ = os.Remove(tmpFileName) }()
	}

	fmt.Println("Saved to", tmpFileName)

	// open collection
	if openWith != "" {
		cmd := exec.Command(openWith, tmpFileName)
		outByts, err := cmd.Output()
		if outByts != nil {
			fmt.Println(string(outByts))
		}
		if err != nil {
			return err
		}
		fmt.Println("<enter> to continue")
		fmt.Scanln()

		byts, err := os.ReadFile(tmpFileName)
		if err != nil {
			return err
		}

		updatedNodes, err := sgf.LoadCollectionSGF(string(byts))
		if err != nil {
			return err
		}

		sgfs = []string{}
		for _, node := range updatedNodes {
			sgfs = append(sgfs, node.SGF())
		}
	}

	return convertCollectionToAnki(sgfs)
}

func convertCollectionToAnki(sgfs []string) error {
	// Read collection and convert to CSV
	var csvRows [][]string
	for _, s := range sgfs {
		node, err := sgf.LoadSGF(s)
		if err != nil {
			return err
		}
		if len(node.Children()) > 0 {
			csvRows = append(csvRows, []string{s, strings.Join(node.AllValues("TAGS"), ",")})
		} else {
			fmt.Println("No children nodes => ignore")
		}
	}

	var writers []*os.File
	if appendToFile != "" {
		fmt.Printf("Appending %d rows to %s\n", len(csvRows), appendToFile)
		f, err := os.OpenFile(appendToFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		writers = append(writers, f)
	}
	if outFile == "" {
		outFile = fmt.Sprintf("anki_%s.csv", time.Now().Format("2006-01-02T15:04:05"))
	}
	if outFile != appendToFile {
		fmt.Printf("Saving %d rows to %s\n", len(csvRows), outFile)
		f, err := os.Create(outFile)
		if err != nil {
			return err
		}
		writers = append(writers, f)
	}

	for _, f := range writers {
		csvwriter := csv.NewWriter(f)
		csvwriter.Comma = ';'
		if err := csvwriter.WriteAll(csvRows); err != nil {
			return err
		}
		csvwriter.Flush()
	}

	return nil
}

func findAnkiNodes(node *sgf.Node, depth int) (ankiNode []*sgf.Node) {
	_, _, ankiLines := parseComment(node)
	if len(node.Children()) == 0 {
		// fmt.Printf("Leaf node at pos %d\n", depth)
		ankiNode = append(ankiNode, node)
	} else if len(ankiLines) > 0 {
		// fmt.Printf("Anki node at pos %d\n", depth)
		ankiNode = append(ankiNode, node)
	}
	for _, child := range node.Children() {
		ankiNode = append(ankiNode, findAnkiNodes(child, depth+1)...)
	}
	return
}
