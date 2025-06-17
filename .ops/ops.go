package main

import (
	"os"

	"labs.lesiw.io/ops/goapp"
	"lesiw.io/ops"
)

func main() {
	goapp.Name = "zync"
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "build")
	}
	ops.Handle(goapp.Ops{})
}
