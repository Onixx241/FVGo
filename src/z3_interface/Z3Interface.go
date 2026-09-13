package z3_interface

import (
	"fmt"
	"main/bmc"
	"main/bmc/nodes"
	"strconv"

	z3 "github.com/Z3Prover/z3/src/api/go"
)

func StateMachine(k int, graph nodes.DesignGraph) {

	ctx := z3.NewContext()

	solver := ctx.NewSolver()

	frameDict := CreateFrameVars(ctx, &graph, k)

	for i := 0; i < k; i++ {

		for _, process := range graph.ProcessIRs {

			for _, assign := range process.Assignments { // maybe I need assertion IOR

				currentState := frameDict[assign.Target.Name][i]
				target := frameDict[assign.Target.Name][i+1]
				val := ExprToZ3(assign.Value, i, frameDict, ctx)
				guard := ExprToZ3(assign.Guard, i, frameDict, ctx)

				eq := ctx.MkEq(target, val) //for my implication -> guards anded and value equals target
				falseEq := ctx.MkEq(target, currentState)

				if guard != nil {

					trueVal := ctx.MkBV(1, 1)
					boolGuard := ctx.MkEq(guard, trueVal)

					implication := ctx.MkImplies(boolGuard, eq)
					falseImplication := ctx.MkImplies(ctx.MkNot(boolGuard), falseEq)

					solver.Assert(implication)
					solver.Assert(falseImplication)

					print(implication.String())

				} else {
					combEq := ctx.MkEq(target, val)
					solver.Assert(combEq)
				}

				_ = val
				_ = guard

			}

		}

	}

	if solver.Check() == z3.Satisfiable {

		if solver.Model() != nil {
			fmt.Print(solver.Model().String())
		}

		print("Satisfiable!")
	}
	if solver.Check() == z3.Unsatisfiable {

		if solver.Model() != nil {
			fmt.Print(solver.Model().String())
		}

		print("Unsatisfiable!")
	}

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

func ExprToZ3(expr *nodes.ExprIR, frame int, frameVars map[string][]*z3.Expr, ctx *z3.Context) *z3.Expr {

	if expr == nil || expr.Kind == "" {
		return nil
	}

	switch expr.Kind {

	case "NamedValue":

		if _, ok := frameVars[expr.Symbol]; ok {

			return frameVars[expr.Symbol][frame]

		} else {

			return nil

		}

	case "IntegerLiteral":
		val := bmc.ParseLiteralValue(expr.Value)
		return ctx.MkBV(val, uint(expr.Width))

	case "Conversion":
		val := bmc.ParseLiteralValue(expr.Left.Value)
		return ctx.MkBV(val, uint(expr.Width))

	case "BinaryOp":
		if expr.Op != "" {

			if expr.Op == "LogicalAnd" {

				left := ExprToZ3(expr.Left, frame, frameVars, ctx)
				right := ExprToZ3(expr.Right, frame, frameVars, ctx)

				implication := ctx.MkBVAnd(left, right)

				fmt.Print(left.String() + "\n")
				fmt.Print(right.String() + "\n")
				fmt.Print(implication.String() + "\n")

				return implication

			}

		}
	case "ConditionalOp":

		iteIf := ExprToZ3(expr.Args[0], frame, frameVars, ctx)
		iteThen := ExprToZ3(expr.Args[1], frame, frameVars, ctx)
		iteElse := ExprToZ3(expr.Args[2], frame, frameVars, ctx)

		trueBv := ctx.MkBV(1, 1)
		falseBv := ctx.MkBV(0, 1)

		z3If := ctx.MkEq(iteIf, trueBv)
		z3False := ctx.MkEq(iteIf, falseBv)

		z3ThenHandle := ctx.MkImplies(z3If, iteThen)
		z3FalseHandle := ctx.MkImplies(z3False, iteElse)

		_ = z3ThenHandle
		_ = z3FalseHandle

	case "UnaryOp":

		if expr.Op != "" {

			switch expr.Op {

			case "LogicalNot":
				left := ExprToZ3(expr.Left, frame, frameVars, ctx)
				return ctx.MkBVNot(left)

			}

		}

	}

	return nil

}
