package main

import (
	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
)

func init() {
	goat.Register(GenerateOverview, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[string]("pkg", "", nil).AsCliFlag(),
			flags.MakeFlag[string]("function", "", nil).AsCliFlag(),
		},
		Name:  "GenerateOverview",
		Usage: "creates a graph overview of the given function and\nprints it out in graphviz DOT format to STDOUT.",
		Action: func(c *cli.Context) error {
			return GenerateOverview(
				flags.GetFlag[string](c, "pkg"),
				flags.GetFlag[string](c, "function"),
			)
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["pkg"] = flags.GetFlag[string](c, "pkg")
			cflags["function"] = flags.GetFlag[string](c, "function")
			return cflags
		},
	})
}
