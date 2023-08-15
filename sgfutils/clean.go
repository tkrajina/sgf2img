package sgfutils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rooklift/sgf"
)

var startRegexp = regexp.MustCompile(`^Move \d+: [WB] \w\d+$`)

func CleanKatrainStuff(node *sgf.Node) error {

	node.DeleteKey("KT")

	comments := node.AllValues(SGFTagComment)
	cleanedComments := []string{}
	katrainLineReached := false
	for _, comment := range comments {
		for _, line := range strings.Split(comment, "\n") {
			if startRegexp.MatchString(strings.TrimSpace(line)) {
				katrainLineReached = true
			}
			if !katrainLineReached {
				cleanedComments = append(cleanedComments, line)
			}
		}
	}
	node.SetValues(SGFTagComment, []string{strings.Join(cleanedComments, "\n")})

	for _, child := range node.Children() {
		if err := CleanKatrainStuff(child); err != nil {
			return err
		}
	}

	return nil
}

func CleanAnkiCommands(node *sgf.Node) error {
	comments := node.AllValues(SGFTagComment)
	cleanedComments := []string{}
	for _, comment := range comments {
		for _, line := range strings.Split(comment, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "!") {
				fmt.Println("Line ignored: ", line)
			} else {
				cleanedComments = append(cleanedComments, line)
			}
		}
	}
	node.SetValues(SGFTagComment, []string{strings.Join(cleanedComments, "\n")})

	for _, child := range node.Children() {
		if err := CleanAnkiCommands(child); err != nil {
			return err
		}
	}

	return nil
}
