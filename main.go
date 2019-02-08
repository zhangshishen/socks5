package main

import (
	"./socks5"
)

func main() {
	p := socks5.SocksProxy{}
	p.Run(":1080")
	return
}
