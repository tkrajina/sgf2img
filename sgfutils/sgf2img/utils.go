package sgf2img

import "strings"

func isOneOf(s string, strs []string) bool {
	for _, str := range strs {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}
