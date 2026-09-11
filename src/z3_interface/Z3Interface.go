package z3_interface

import (
	"fmt"

	z3 "github.com/Z3Prover/z3/src/api/go"
)

func Test() {

	ctx := z3.NewContext()
	solver := ctx.NewSolver()

	x := ctx.MkBV(1, 32)
	y := ctx.MkBV(1, 32)

	eq := ctx.MkEq(x, y)

	solver.Assert(eq)

	if solver.Check() == z3.Satisfiable {
		fmt.Println("Satisfiable.")
	} else if solver.Check() == z3.Unsatisfiable {
		fmt.Println("Unsatisfiable")
	}

}
