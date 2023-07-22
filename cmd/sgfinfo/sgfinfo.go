package main

import (
	"flag"
	"fmt"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

var args struct {
	all      bool
	comments bool
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.BoolVar(&args.all, "all", false, "All informations")
	flag.BoolVar(&args.comments, "c", false, "Print comments")
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			return err
		}

		gi := sgfutils.ParseGameInfo(node)
		if args.all {
			fmt.Println("File:", fn)
			fmt.Println("Date:", gi.Date)
			fmt.Println("Event:", gi.Event)
			fmt.Println("BlackName:", gi.BlackName)
			fmt.Println("BlackRank:", gi.BlackRank)
			fmt.Println("BlackTeam:", gi.BlackTeam)
			fmt.Println("WhiteName:", gi.WhiteName)
			fmt.Println("WhiteRank:", gi.WhiteRank)
			fmt.Println("WhiteTeam:", gi.WhiteTeam)
			fmt.Println("Result:", gi.Result)
			fmt.Println("Rules:", gi.Rules)
			fmt.Println("Komi:", gi.Komi)
			fmt.Println("Handicap:", gi.Handicap)
			fmt.Println("----------------------------------------------------------------------------------------------------")
		} else {
			fmt.Printf("%20s %5s vs %20s %5s: %s\n", gi.BlackName, gi.BlackRank, gi.WhiteName, gi.WhiteRank, fn)
		}

		if args.comments {
			if err := printComments(node, 1); err != nil {
				return err
			}
		}
	}
	return nil
}

func printComments(node *sgf.Node, depth int) error {
	comments := node.AllValues(sgfutils.SGFTagComment)
	for n, comment := range comments {
		fmt.Printf("move #%d comment #%d: %s\n", depth, n+1, comment)
	}

	for _, child := range node.Children() {
		if err := printComments(child, depth+1); err != nil {
			return err
		}
	}

	return nil
}
