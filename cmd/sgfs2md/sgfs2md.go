package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/rooklift/sgf"
	"github.com/tkrajina/sgf2img/sgfutils"
	"github.com/tkrajina/sgf2img/sgfutils/sgf2img"

	"github.com/gomarkdown/markdown"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var recursive bool
var html bool
var opts = sgf2img.Options{AutoCrop: true, ImageSize: 150, ImageType: sgf2img.SVG, Images: []int{0, 1}}

func main() {
	flag.BoolVar(&recursive, "r", false, "Recursive from current from directories")
	flag.BoolVar(&html, "h", false, "Convert to html")
	flag.Parse()

	b := bytes.NewBufferString("")

	var lastDir string
	for _, fn := range flag.Args() {
		if recursive {
			walk(fn, func(filename string) error {
				dir, fn := path.Split(filename)
				_ = fn
				fmt.Println(dir, lastDir)
				if dir != lastDir {
					parts := strings.Split(dir, string(os.PathSeparator))
					var title []string
					for _, part := range parts {
						part = strings.Trim(part, ". \t\r\n")
						if len(part) > 0 {
							title = append(title, part)
						}
					}
					_, _ = b.WriteString("## " + strings.ReplaceAll(strings.Join(title, " / "), "_", " ") + "\n\n")
				}
				lastDir = dir
				if strings.HasSuffix(strings.ToLower(filename), ".sgf") {
					panicIfErr(file(filename, b))
				}
				return nil
			})
		} else {
			panicIfErr(file(fn, b))
		}
	}

	if html {
		output := markdown.ToHTML(b.Bytes(), nil, nil)
		output = append([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>SGFs</title>
    <meta charset="utf-8">
</head>
<body>
		`), output...)
		output = append(output, []byte("\n</html>")...)
		panicIfErr(os.WriteFile("sgfs.html", output, 0700))
	} else {
		panicIfErr(os.WriteFile("sgfs.md", b.Bytes(), 0700))
	}
}

func file(fn string, b *bytes.Buffer) error {
	node, images, err := sgf2img.ProcessSgfFile(fn, &opts)
	if err != nil {
		return err
	}
	_ = node

	if len(images) == 0 {
		return fmt.Errorf("no images found for %s", fn)
	}

	_, _ = b.WriteString(fmt.Sprintf(`<div style="width: %dpx;">`, opts.ImageSize))
	_, _ = b.Write(images[0].Contents)
	_, _ = b.WriteString("</div>\n")

	_, _ = b.WriteString(fmt.Sprintf("\n\n**%s:** ", strings.Replace(path.Base(fn), ".sgf", "", 1)))
	comments := node.AllValues(sgfutils.SGFTagComment)
	if len(comments) > 0 {
		_, _ = b.WriteString(comments[0])
	}
	_, _ = b.WriteString("<br/>")

	solution := getURL(node)
	//_, _ = b.WriteString(nodeToSGF(node) + "\n\n")
	for _, sub := range node.Children() {
		sub.Detach()
	}
	problem := getURL(node)
	//_, _ = b.WriteString(nodeToSGF(node) + "\n\n")

	_, _ = b.WriteString("[*problem*](" + problem + ") ")
	_, _ = b.WriteString("· [*solution*](" + solution + ") \n\n")

	return nil
}

func walk(dir string, onFile func(fn string) error) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	var (
		filenames   []string
		directories []string
	)
	for _, fn := range files {
		if fn.IsDir() {
			directories = append(directories, fn.Name())
		} else {
			filenames = append(filenames, fn.Name())
		}
	}

	sort.Strings(directories)
	sort.Strings(filenames)

	for _, subdir := range directories {
		walk(path.Join(dir, subdir), onFile)
	}

	for _, fn := range filenames {
		if err := onFile(path.Join(dir, fn)); err != nil {
			return err
		}
	}

	return nil
}

func getURL(node *sgf.Node) string {
	fn := os.TempDir() + string(os.PathSeparator) + "tmp_" + fmt.Sprint(time.Now().Nanosecond()) + ".sgf"
	panicIfErr(node.Save(fn))

	byts, err := os.ReadFile(fn)
	panicIfErr(err)

	v := url.Values{}
	v.Set("sgf", string(byts))
	return "https://tkrajina.github.io/besogo/share.html?" + v.Encode()
}
