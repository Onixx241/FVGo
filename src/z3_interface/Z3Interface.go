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

			for _, assign := range process.Assignments { // maybe I need assertion IR

				currentState := frameDict[assign.Target.Name][i]
				target := frameDict[assign.Target.Name][i+1]
				val := ExprToZ3(assign.Value, i, frameDict, ctx, solver)
				guard := ExprToZ3(assign.Guard, i, frameDict, ctx, solver)

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

	// for _, assertion := range graph.AssertionIRs {

	// 	var failureConditions []*z3.Expr

	// 	for i := 0; i < k; i++ {

	// 		condition := AssertionToZ3(assertion, i, frameDict, ctx, solver)

	// 		trueBool := ctx.MkBV(1, 1)

	// 		implication := ctx.MkImplies(condition, trueBool)

	// 		solver.Assert(implication)

	// 	}

	// 	if len(failureConditions) > 0 {

	// 	}

	// }

	if solver.Check() == z3.Satisfiable {

		if solver.Model() != nil {
			fmt.Print(solver.Model().String())
		}

		print("Satisfiable!")
	}
	if solver.Check() == z3.Unsatisfiable {

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

func AssertionToZ3(assertion *nodes.AssertionIR, frame int, frameVars map[string][]*z3.Expr, ctx *z3.Context, solver *z3.Solver) *z3.Expr {

	if assertion == nil || assertion.Implication == "" {
		return nil
	}

	// antecedent := ExprToZ3(assertion.Antecedent.Left, frame, frameVars, ctx, solver)
	// consequent := ExprToZ3(assertion.Consequent.Left)

	return nil
}

func ExprToZ3(expr *nodes.ExprIR, frame int, frameVars map[string][]*z3.Expr, ctx *z3.Context, solver *z3.Solver) *z3.Expr {

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

				left := ExprToZ3(expr.Left, frame, frameVars, ctx, solver)
				right := ExprToZ3(expr.Right, frame, frameVars, ctx, solver)

				implication := ctx.MkBVAnd(left, right)

				fmt.Print(left.String() + "\n")
				fmt.Print(right.String() + "\n")
				fmt.Print(implication.String() + "\n")

				return implication

			}

		}
	case "ConditionalOp":

		iteIf := ExprToZ3(expr.Args[0], frame, frameVars, ctx, solver)
		iteThen := ExprToZ3(expr.Args[1], frame, frameVars, ctx, solver)
		iteElse := ExprToZ3(expr.Args[2], frame, frameVars, ctx, solver)

		return ResolveConditionalITE(ctx, solver, expr, iteIf, iteThen, iteElse, frame)

	case "Equality":
		panic("implement next for assertions!")

	case "UnaryOp":

		if expr.Op != "" {

			switch expr.Op {

			case "LogicalNot":
				left := ExprToZ3(expr.Left, frame, frameVars, ctx, solver)
				return ctx.MkBVNot(left)

			}

		}

	}

	return nil

}

func ResolveConditionalITE(ctx *z3.Context, solver *z3.Solver, condExpr *nodes.ExprIR, ifExpr *z3.Expr, thenExpr *z3.Expr, elseExpr *z3.Expr, frame int) *z3.Expr {

	tempName := fmt.Sprintf("ite_tmp_%p_f%d", condExpr, frame)
	tempVar := ctx.MkConst(ctx.MkStringSymbol(tempName), ctx.MkBvSort(uint(condExpr.Width)))

	trueBv := ctx.MkBV(1, 1)

	z3If := ctx.MkEq(ifExpr, trueBv)
	z3False := ctx.MkNot(z3If)

	thenEq := ctx.MkEq(tempVar, thenExpr)
	elseEq := ctx.MkEq(tempVar, elseExpr)

	solver.Assert(ctx.MkImplies(z3If, thenEq))
	solver.Assert(ctx.MkImplies(z3False, elseEq))

	return tempVar

}
