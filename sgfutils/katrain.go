package sgfutils

import (
	"regexp"
	"strings"

	"github.com/rooklift/sgf"
)

var startRegexp = regexp.MustCompile(`^Move \d+: [WB] \w\d+$`)

func CleanKatrainStuff(node *sgf.Node) error {

	node.DeleteKey("KT")

	comments := node.AllValues(SGFTagComment)
	cleanedComments := []string{}
	ankiLlineReached := false
	for _, comment := range comments {
		for _, line := range strings.Split(comment, "\n") {
			if startRegexp.MatchString(strings.TrimSpace(line)) {
				ankiLlineReached = true
			}
			if !ankiLlineReached {
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
