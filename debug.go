package qsim

import (
	"fmt"
)

// Debug determines whether debug output will be displayed
var Debug bool

// D writes the given debug output to stdout if debug output is enabled.
func D(a ...interface{}) {
	if Debug {
		fmt.Println(a...)
	}
}
