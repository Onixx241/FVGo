package z3_interface

import (
	"fmt"
	"main/bmc/nodes"
	"strconv"

	z3 "github.com/Z3Prover/z3/src/api/go"
)

func Test() {

	ctx := z3.NewContext()
	solver := ctx.NewSolver()

	x := ctx.MkBV(5, 32)
	y := ctx.MkBV(5, 32)

	eq := ctx.MkEq(x, y)

	solver.Assert(eq)

	if solver.Check() == z3.Satisfiable {
		fmt.Println("Satisfiable.")
	} else if solver.Check() == z3.Unsatisfiable {
		fmt.Println("Unsatisfiable")
	}

}

func Test2(graph *nodes.DesignGraph) { //might need to edit graph ill leave pointer for now

	ctx := z3.NewContext()
	solver := ctx.NewSolver()
	bitVectors := []*z3.Expr{}

	for _, o := range graph.AssertionIRs {

		temp := ctx.MkBV(o.Antecedent.Left.Width, uint(o.Antecedent.Left.Width))
		bitVectors = append(bitVectors, temp)

	}

	solver.Check()

}

func CreateFrameVars(ctx *z3.Context, graph *nodes.DesignGraph, k int) map[string][]*z3.Expr {

	exprMap := make(map[string][]*z3.Expr)

	add := func(name string, width int) {

		if _, ok := exprMap[name]; ok {
			return
		}

		for t := 0; t <= k; t++ {

			sym := ctx.MkStringSymbol(name + "_" + strconv.Itoa(t))

			v := ctx.MkConst(sym, ctx.MkBvSort(uint(width)))

			exprMap[name] = append(exprMap[name], v)

		}

	}

	for _, s := range graph.Inputs {
		add(s.Name, s.Width)
	}
	for _, s := range graph.Signals {
		add(s.Name, s.Width)
	}
	for _, s := range graph.Outputs {
		add(s.Name, s.Width)
	}

	return exprMap
}
