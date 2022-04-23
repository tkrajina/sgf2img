package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
)

func walkNodesAndMarkMistakes(node *sgf.Node, opts *ctx, depth int) error {
	comment, _ := node.GetValue(sgfutils.SGFTagComment)
	c := strings.TrimSpace(strings.ToLower(comment))
	if strings.HasPrefix(c, "mistake") && !strings.HasPrefix(c, "mistakes") {
		fmt.Println("Mistake node:", comment)
		markMistakes(node)
	}

	for _, child := range node.Children() {
		if err := walkNodesAndMarkMistakes(child, opts, depth+1); err != nil {
			return err
		}
	}

	return nil
}

func markMistakes(mistakeNode *sgf.Node) {
	parentNode := mistakeNode.Parent()
	if parentNode == nil {
		fmt.Println("Mistake node without parent node")
		fmt.Println(sgfutils.BoardToString(*mistakeNode.Board()))
		return
	}
	if len(parentNode.Children()) != 2 {
		fmt.Println("Mistake node parent node has no 2 branches")
		fmt.Println(sgfutils.BoardToString(*mistakeNode.Board()))
		return
	}
	id := fmt.Sprint(time.Now().UnixNano())
	comment, _ := parentNode.GetValue(sgfutils.SGFTagComment)
	parentNode.SetValue(sgfutils.SGFTagComment, directiveImg[0]+" "+id+"\n"+comment)
	fmt.Println("Setting start ", id)

	branch := parentNode.Children()[1]
	for len(branch.Children()) > 0 {
		branch = branch.Children()[0]
	}

	comment, _ = branch.GetValue(sgfutils.SGFTagComment)
	branch.SetValue(sgfutils.SGFTagComment, directiveEnd[0]+" "+id+"\n"+comment)
	fmt.Println("Setting end ", id)
}
