package sgfutils

import (
	"strings"

	"github.com/rooklift/sgf"
)

type GameInfo struct {
	RootNode *sgf.Node
	Date     string

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

	SuggestedFilename string
}

func ParseGameInfo(node *sgf.Node) GameInfo {
	res := GameInfo{RootNode: node}
	res.Date, _ = node.GetValue(SGFTagDate)

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

	res.SuggestedFilename = strings.ReplaceAll(strings.Join(newFn, "_"), " ", "_")
	return res
}
