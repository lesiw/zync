package main

import (
	"os"

	"labs.lesiw.io/ops/goapp"
	"labs.lesiw.io/ops/golang"
	"lesiw.io/ops"
)

func main() {
	goapp.Name = "zync"
	goapp.Targets = []golang.Target{
		{Goos: "linux", Goarch: "386"},
		{Goos: "linux", Goarch: "amd64"},
		{Goos: "linux", Goarch: "arm"},
		{Goos: "linux", Goarch: "arm64"},
		{Goos: "darwin", Goarch: "amd64"},
		{Goos: "darwin", Goarch: "arm64"},
	}
	golang.CheckTargets = []golang.Target{
		{Goos: "linux", Goarch: "386"},
		{Goos: "linux", Goarch: "amd64"},
		{Goos: "linux", Goarch: "arm"},
		{Goos: "linux", Goarch: "arm64"},
		{Goos: "darwin", Goarch: "amd64"},
		{Goos: "darwin", Goarch: "arm64"},
	}
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "check")
	}
	ops.Handle(goapp.Ops{})
}
