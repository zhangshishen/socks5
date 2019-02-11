package socks5

import (
	"fmt"
)

type Debug struct {
}

var debug Debug

func (d *Debug) output(s string, a ...interface{}) {
	fmt.Printf(s, a...)
}

func (d *Debug) out(s string, a ...interface{}) {
	fmt.Printf(s, a...)
}
