package sgfutils

import (
	"strings"
	"unicode"

	"github.com/rooklift/sgf"
)

type GameInfo struct {
	RootNode *sgf.Node
	Date     string

	Event string

	BlackName string
	BlackRank string
	BlackTeam string

	WhiteName string
	WhiteRank string
	WhiteTeam string

	Result   string
	Rules    string
	Komi     string
	Handicap string

	Moves int

	SuggestedFilename string
}

func ParseGameInfo(node *sgf.Node) GameInfo {
	res := GameInfo{RootNode: node}
	res.Date, _ = node.GetValue(SGFTagDate)
	res.Event, _ = node.GetValue(SGFTagEvent)

	res.BlackName, _ = node.GetValue(SGFTagBlackName)
	res.BlackTeam, _ = node.GetValue(SGFTagBlackTeam)
	res.BlackRank, _ = node.GetValue(SGFTagBlackRank)

	res.WhiteName, _ = node.GetValue(SGFTagWhiteName)
	res.WhiteTeam, _ = node.GetValue(SGFTagWhiteTeam)
	res.WhiteRank, _ = node.GetValue(SGFTagWhiteRank)

	res.Result, _ = node.GetValue(SGFTagResult)
	res.Rules, _ = node.GetValue(SGFTagRules)
	res.Komi, _ = node.GetValue(SGFTagKomi)
	res.Handicap, _ = node.GetValue(SGFTagHandicap)

	tmpNode := node
	for {
		tmpNode = tmpNode.MainChild()
		if tmpNode == nil {
			break
		} else {
			res.Moves++
		}
	}

	var newFn []string
	if res.Date != "" {
		newFn = append(newFn, res.Date)
	}

	if res.BlackName == "" {
		newFn = append(newFn, "NONAME")
	} else {
		newFn = append(newFn, res.BlackName)
	}
	if res.BlackRank != "" {
		newFn = append(newFn, res.BlackRank)
	}
	newFn = append(newFn, "vs")

	if res.WhiteName == "" {
		newFn = append(newFn, "NONAME")
	} else {
		newFn = append(newFn, res.WhiteName)
	}
	if res.BlackRank != "" {
		newFn = append(newFn, res.WhiteRank)
	}

	for n := range newFn {
		cleared := ""
		for _, r := range []rune(newFn[n]) {
			if unicode.IsSpace(r) {
				cleared += string(" ")
			} else if strings.ToLower(string(r)) != strings.ToUpper(string(r)) {
				cleared += string(r)
			} else if unicode.IsDigit(r) {
				cleared += string(r)
			} else if strings.ContainsRune("().,-", r) {
				cleared += string(r)
			}
		}
		newFn[n] = strings.TrimSpace(cleared)
	}

	res.SuggestedFilename = strings.ReplaceAll(strings.Join(newFn, "_"), " ", "_")
	res.SuggestedFilename = strings.ReplaceAll(res.SuggestedFilename, "__", "_")
	return res
}
