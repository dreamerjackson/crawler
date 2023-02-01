package main

import (
	"github.com/dreamerjackson/crawler/cmd"
	_ "github.com/dreamerjackson/crawler/tasklib"
	_ "net/http/pprof"
)

func main() {
	cmd.Execute()
}
