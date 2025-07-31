package main

import (
	_ "net/http/pprof"

	"github.com/dreamerjackson/crawler/cmd"
)

func main() {
	cmd.Execute()
}
