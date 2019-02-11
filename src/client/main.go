package main

import "../socks5"

func main() {
	p := socks5.SocksProxy{}

	p.RunRelay(":1080", ":1090")
	return
}
