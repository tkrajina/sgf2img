package main

import (
	"flag"
	"fmt"
	"path"

	"github.com/rooklift/sgf"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()
	panicIfErr(doStuff())
}

func doStuff() error {
	for _, fn := range flag.Args() {
		node, err := sgf.Load(fn)
		if err != nil {
			return err
		}

		candidate := leafNodeCandidate{depth: 0, node: node}
		walkNodes(node, 0, &candidate)

		dir, file := path.Split(fn)
		newFn := path.Join(dir, "longest_mainline_"+file)
		candidate.node.MakeMainLine()
		panicIfErr(candidate.node.Save(newFn))
		fmt.Println("Saved to", newFn)
	}

	return nil
}

type leafNodeCandidate struct {
	depth int
	node  *sgf.Node
}

func walkNodes(node *sgf.Node, depth int, leafNode *leafNodeCandidate) {
	if len(node.Children()) == 0 {
		fmt.Printf("Leaf node at pos %d\n", depth)
		if depth > leafNode.depth {
			fmt.Printf("Longest branch node candidate at pos %d\n", depth)
			leafNode.node = node
			leafNode.depth = depth
		}
	}
	for _, child := range node.Children() {
		walkNodes(child, depth+1, leafNode)
	}
}
