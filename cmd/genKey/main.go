package main

import (
	"fmt"
	"os"

	"git.taiyue.io/pist/go-pist/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

// Git SHA1 commit hash of the release (set via linker flags)
var gitCommit = ""
var gitData = ""
var app *cli.App

func init() {
	app = utils.NewApp(gitCommit, gitData, "an pist generate key tool")
	app.Commands = []cli.Command{
		commandGenerate,
	}
	app.Flags = append(app.Flags, commandGenerate.Flags...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
