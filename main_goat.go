package main

import (
	"github.com/tmr232/goat"
	"github.com/tmr232/goat/flags"
	"github.com/urfave/cli/v2"
)

func init() {
	goat.Register(GenerateOverview, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[string]("pkg", "The path of the package to load.\nYou may need to run 'go get `package`' to fetch it first.", nil).AsCliFlag(),
			flags.MakeFlag[string]("function", "The name of the function to generate an overview of.", nil).AsCliFlag(),
		},
		Name:  "function",
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

	goat.Register(PackageOverview, goat.RunConfig{
		Flags: []cli.Flag{
			flags.MakeFlag[string]("pkg", "The path of the package to load.\nYou may need to run 'go get `package`' to fetch it first.", nil).AsCliFlag(),
			flags.MakeFlag[string]("out", "Output file will be written to `path`.", nil).AsCliFlag(),
		},
		Name:  "package",
		Usage: "generates an overview for an entire package.",
		Action: func(c *cli.Context) error {
			return PackageOverview(
				flags.GetFlag[string](c, "pkg"),
				flags.GetFlag[string](c, "out"),
			)
		},
		CtxFlagBuilder: func(c *cli.Context) map[string]any {
			cflags := make(map[string]any)
			cflags["pkg"] = flags.GetFlag[string](c, "pkg")
			cflags["out"] = flags.GetFlag[string](c, "out")
			return cflags
		},
	})
}
