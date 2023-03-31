package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/utils"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

type row struct {
	Id   int64    `json:"id"`
	Flds []string `json:"flds"`
}

var ignoreRegexps = []*regexp.Regexp{
	regexp.MustCompile(`<\w+.*?>`),
	regexp.MustCompile(`<\/\w+.*?>`),
}

var newlineRegexp = []*regexp.Regexp{
	regexp.MustCompile(`<div.*?>`),
	regexp.MustCompile(`<br.*?>`),
	regexp.MustCompile(`<p.*?>`),
}

func main() {
	panicIfErr(doStuff())
}

func doStuff() error {
	flag.Parse()

	ankiFile := flag.Arg(0)
	if ankiFile == "" {
		fmt.Fprint(os.Stderr, "Missing anki collection argument")
	}
	sgfFile := flag.Arg(1)
	if sgfFile == "" {
		fmt.Println("Paste SGF here (^D to exit):")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		if scanner.Err() != nil {
			return scanner.Err()
		}

		println(string(scanner.Bytes()))
	}

	db, err := sql.Open("sqlite3", ankiFile+"?mode=ro")
	if err != nil {
		return err
	}

	// rows, err := db.Query("select id, flds, sfld from notes where mid=(select id from notetypes n where name like 'goban') and id=1659089922340")
	rows, err := db.Query("SELECT id, flds, sfld FROM notes WHERE flds like '%tkrajina%'")
	if err != nil {
		return err
	}

	var all []row

	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
		if n%100 == 0 {
			fmt.Printf("Parsing sgf #%d\n", n)
		}
		var (
			id         int64
			flds, sfld string
		)
		err := rows.Scan(&id, &flds, &sfld)
		fmt.Println(flds)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("flds", flds)
		fmt.Printf("sfld", sfld)
		parts := strings.Split(flds, string(rune(31)))
		//fmt.Printf("%d %#v\n", id, parts)
		all = append(all, row{
			Id:   id,
			Flds: parts,
		})

		s := parts[0]

		for _, r := range newlineRegexp {
			s = r.ReplaceAllLiteralString(s, "\n")
		}
		for _, r := range ignoreRegexps {
			s = r.ReplaceAllLiteralString(s, "")
		}

		parsed, err := sgf.LoadSGF(s)
		if err != nil {
			runes := []rune(s)
			min, _ := utils.MinMax(100, len(runes))
			fmt.Printf("nid:%d, error parsing: %s...\n", id, string(runes[:min]))
		}
		_ = parsed
	}

	return nil
}
