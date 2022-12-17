package main

import (
	"github.com/dreamerjackson/crawler/cmd"
	_ "net/http/pprof"
)

func main() {
	cmd.Execute()
}
