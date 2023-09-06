package main

import (
	"github.com/elwin/franz/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.BuildVersion = cmd.Version{
		Tag:    version,
		Commit: commit,
		Date:   date,
	}
	cmd.Execute()
}
