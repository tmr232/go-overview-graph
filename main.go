package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tmr232/goat"
	"go/token"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
	"io/ioutil"
	"log"
)

type Node struct {
	index    int
	function *ssa.Function
}

func (n Node) Attributes() []encoding.Attribute {
	var lines []string
	block := n.function.Blocks[n.index]
	for _, instr := range block.Instrs {
		lines = append(lines, instr.String())
	}

	fillColor := "gray"
	shape := "box"
	minHeight := 0.2
	if len(block.Succs) == 0 && len(block.Preds) != 0 {
		fillColor = "#AB3030" // Dark red
		shape = "house"
		minHeight = 0.5
	} else if len(block.Succs) != 0 && len(block.Preds) == 0 {
		fillColor = "#48AB30" // Dark green
		shape = "invhouse"
		minHeight = 0.5
	}

	return []encoding.Attribute{
		//{"label", strings.Join(lines, "\n") + "\n"},
		{"label", ""},
		{"shape", shape},
		{"fixedsize", "true"},
		{"height", fmt.Sprint(floats.Max([]float64{minHeight, float64(len(block.Instrs)) * 0.1}))},
		{"fillcolor", fillColor},
		{"style", "filled"},
	}
}

func (n Node) ID() int64 {
	return int64(n.index)
}

type Edge struct {
	from     int
	to       int
	function *ssa.Function
}

func (e Edge) FromPort() (port, compass string) {
	return "", "s"
}

func (e Edge) ToPort() (port, compass string) {
	return "", "n"
}

func (e Edge) Attributes() []encoding.Attribute {
	color := "blue"
	from := e.function.Blocks[e.from]
	to := e.function.Blocks[e.to]
	if len(from.Succs) == 2 {
		if to == from.Succs[0] {
			color = "green"
		} else {
			color = "red"
		}
	}
	penWidth := "1.0"
	if e.from > e.to {
		penWidth = "2.0"
	}
	return []encoding.Attribute{
		{"color", color},
		{"penwidth", penWidth},
	}
}

func (e Edge) From() graph.Node {
	return Node{e.from, e.function}
}

func (e Edge) To() graph.Node {
	return Node{e.to, e.function}
}

func (e Edge) ReversedEdge() graph.Edge {
	return Edge{e.to, e.from, e.function}
}

func blocksToDot(function *ssa.Function) ([]byte, error) {
	functionGraph := simple.NewDirectedGraph()

	for _, block := range function.Blocks {
		functionGraph.AddNode(&Node{index: block.Index, function: function})
	}
	for _, block := range function.Blocks {
		for _, succ := range block.Succs {
			functionGraph.SetEdge(Edge{
				from:     block.Index,
				to:       succ.Index,
				function: function,
			})
		}
	}

	dotGraph, err := dot.Marshal(functionGraph, function.Name(), "", "    ")
	dotGraph = bytes.Replace(dotGraph, []byte("\\n"), []byte("\\l"), -1)
	return dotGraph, err
}

type entry struct {
	Func     *ssa.Function
	Position token.Position
}

type FunctionOverview struct {
	Name string `json:"name"`
	Dot  string `json:"dot"`
	Line int    `json:"line"`
}

type FileOverview struct {
	Filename  string             `json:"filename"`
	Functions []FunctionOverview `json:"functions"`
}

type Overview map[string]*FileOverview

// PackageOverview generates an overview for an entire package.
func PackageOverview(pkg string, outpath string) error {
	goat.Self().Name("package")
	goat.Flag(pkg).Usage("The path of the package to load.\nYou may need to run 'go get `package`' to fetch it first.")
	goat.Flag(outpath).Name("out").Usage("Output file will be written to `path`.")

	// Load, parse, and type-check the initial packages.
	cfg := &packages.Config{Mode: packages.LoadSyntax}
	initial, err := packages.Load(cfg, pkg)
	if err != nil {
		return err
	}

	// Stop if any package had errors.
	// This step is optional; without it, the next step
	// will create SSA for only a subset of packages.
	if packages.PrintErrors(initial) > 0 {
		log.Fatalf("packages contain errors")
	}

	// Create SSA packages for all well-typed packages.
	prog, pkgs := ssautil.Packages(initial, 0)
	_ = prog

	// Build SSA code for the well-typed initial packages.
	for _, p := range pkgs {
		if p != nil {
			p.Build()
		}
	}

	overview := make(Overview)
	for f, _ := range ssautil.AllFunctions(prog) {
		if f.Pkg != pkgs[0] {
			continue
		}
		pos := prog.Fset.Position(f.Pos())

		fileOverview, exists := overview[pos.Filename]
		if !exists {
			fileOverview = &FileOverview{Filename: pos.Filename}
			overview[pos.Filename] = fileOverview
		}
		funcDot, err := blocksToDot(f)
		if err != nil {
			return errors.Wrap(err, "Failed converting function to dot")
		}
		fileOverview.Functions = append(fileOverview.Functions, FunctionOverview{
			Name: f.Name(),
			Dot:  string(funcDot),
			Line: pos.Line,
		})
	}

	jsonData, err := json.Marshal(overview)
	if err != nil {
		return errors.Wrap(err, "failed marshaling JSON")
	}

	err = ioutil.WriteFile(outpath, jsonData, 0o666)
	if err != nil {
		return errors.Wrap(err, "Failed writing result")
	}
	return nil
}

// GenerateOverview creates a graph overview of the given function and
// prints it out in graphviz DOT format to STDOUT.
func GenerateOverview(pkg string, function string, png *string) error {
	goat.Self().Name("function")
	goat.Flag(pkg).Usage("The path of the package to load.\nYou may need to run 'go get `package`' to fetch it first.")
	goat.Flag(function).Usage("The name of the function to generate an overview of.")
	goat.Flag(png).Usage("An optional `path` to save a rendered PNG to.\nWhen used, DOT will not be printed to STDOUT.")

	// Load, parse, and type-check the initial packages.
	cfg := &packages.Config{Mode: packages.LoadSyntax}
	initial, err := packages.Load(cfg, pkg)
	if err != nil {
		return err
	}

	// Stop if any package had errors.
	// This step is optional; without it, the next step
	// will create SSA for only a subset of packages.
	if packages.PrintErrors(initial) > 0 {
		log.Fatalf("packages contain errors")
	}

	// Create SSA packages for all well-typed packages.
	prog, pkgs := ssautil.Packages(initial, 0)
	_ = prog

	// Build SSA code for the well-typed initial packages.
	for _, p := range pkgs {
		if p != nil {
			p.Build()
		}
	}

	for _, p := range pkgs {
		ssaFunc := p.Func(function)
		if ssaFunc == nil {
			return fmt.Errorf("function %s not found", function)
		}
		data, err := blocksToDot(ssaFunc)
		if err != nil {
			return err
		}
		if png == nil {
			fmt.Println(string(data))
		} else {
			err = renderPng(data, *png)
			if err != nil {
				return errors.Wrap(err, fmt.Sprint("Failed rendering PNG to ", png))
			}
		}
	}

	return nil
}

//go:generate go run github.com/tmr232/goat/cmd/goater
func main() {
	goat.App("graph-overview",
		goat.Command(GenerateOverview),
		goat.Command(PackageOverview),
	).Run()
}
