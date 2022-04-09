package sgfutils

import "fmt"

func Input(msg string) string {
	fmt.Print(msg + ":")
	var target string
	fmt.Scanln(&target)
	return target
}
