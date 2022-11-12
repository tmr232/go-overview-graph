package main

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
	"log"
	"os"
)

func Demo(flag bool) {
	if flag {
		fmt.Println("Hello, World!")
	}
	fmt.Println("Oh no!")

	for _, x := range "Hello, World!" {
		fmt.Println(x)
	}
}

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

func main() {
	// Load, parse, and type-check the initial packages.
	cfg := &packages.Config{Mode: packages.LoadSyntax}
	initial, err := packages.Load(cfg, ".")
	if err != nil {
		log.Fatal(err)
	}

	// Stop if any package had errors.
	// This step is optional; without it, the next step
	// will create SSA for only a subset of packages.
	if packages.PrintErrors(initial) > 0 {
		log.Fatalf("packages contain errors")
	}

	// Create SSA packages for all well-typed packages.
	prog, pkgs := ssautil.Packages(initial, ssa.PrintPackages)
	_ = prog

	// Build SSA code for the well-typed initial packages.
	for _, p := range pkgs {
		if p != nil {
			p.Build()
		}
		p.Func("Demo").WriteTo(os.Stdout)
		for _, block := range p.Func("Demo").Blocks {
			fmt.Println(block.Index, len(block.Instrs))
		}
		data, err := blocksToDot(p.Func("main"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))
	}

	//graph := simple.NewDirectedGraph()
	//graph.AddNode(&Node{0, "abc"})
	//dotGraph, err := dot.Marshal(graph, "Graph", "", "    ")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(string(dotGraph))
}
