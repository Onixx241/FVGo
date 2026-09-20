package z3_interface

import (
	"fmt"
	"log"
	"main/bmc"
	"main/bmc/nodes"
	"strconv"

	z3 "github.com/Z3Prover/z3/src/api/go"
)

const (
	EmptyString = ""
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

	for _, assumption := range graph.AssumptionIRs {
		ConstrainAssumptions(assumption, k, frameDict, ctx, solver)
	}

	for _, assertion := range graph.AssertionIRs {

		var failureConditions []*z3.Expr

		for i := 0; i < k; i++ {

			condition := AssertionToZ3(assertion, i, k, frameDict, ctx, solver)

			if condition != nil {
				failureConditions = append(failureConditions, condition)
			}

		}

		if len(failureConditions) > 0 {

			orAsserts := ctx.MkOr(failureConditions...)

			solver.Assert(orAsserts)

		}

	}

	if solver.Check() == z3.Satisfiable {

		if solver.Model() != nil {
			fmt.Print(solver.Model().String())
		}

		print("Assertion Violation Found!")
		//try to find waveform library to generate waveform files
	}
	if solver.Check() == z3.Unsatisfiable {

		print("No Assertion Violations Found!")

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

func ConstrainAssumptions(assumption *nodes.AssumptionIR, limit int, frameVars map[string][]*z3.Expr, ctx *z3.Context, solver *z3.Solver) {

	sliceItem, exists := frameVars[assumption.ConstraintSymbol]

	if !exists {
		return
	}

	var assumpValue *z3.Expr

	//change to switch later
	if assumption.AssumptionExpr.Op == "LogicalNot" {

		assumpValue = ctx.MkBVNot(sliceItem[0])

	} else if assumption.AssumptionExpr.Op == "Equality" {

		assumpValue = ctx.MkBV(bmc.ParseLiteralValue(assumption.AssumptionExpr.Value), uint(assumption.AssumptionExpr.Width))

	} else {

		return //add others later

	}

	for i := 0; i < limit; i++ {

		targ := sliceItem[i]
		solver.Assert(ctx.MkEq(targ, assumpValue))

	}

}

func AssertionToZ3(assertion *nodes.AssertionIR, frame int, limit int, frameVars map[string][]*z3.Expr, ctx *z3.Context, solver *z3.Solver) *z3.Expr {

	//if non overlapped implication use min max
	if assertion == nil || assertion.Implication == "" {
		return nil
	}

	trueBv := ctx.MkBV(1, 1)

	delay := 0

	if assertion.DelayMax != 0 || assertion.DelayMin != 0 {

		if frame+assertion.DelayMax > limit {

			boundViolatedAssertionMessage := RecurseLeftsForSymbol(assertion.Antecedent) + "->" + RecurseLeftsForSymbol(assertion.Consequent) + "\n\nCurrent Frame: " + strconv.FormatInt(int64(frame), 10) + "\n\nDelay: " + strconv.FormatInt(int64(assertion.DelayMax), 10) + "\n\n" + "K-Bound Maximum: " + strconv.FormatInt(int64(limit), 10)

			log.Print("\n\nOne of your temporal assertions has a clock delay bigger than the k bound!\n\n", boundViolatedAssertionMessage+"\n\n")

			return nil

		} else {

			delay = assertion.DelayMax

		}

	}

	antecedent := ExprToZ3(assertion.Antecedent, frame, frameVars, ctx, solver)

	consequent := ExprToZ3(assertion.Consequent, frame+delay, frameVars, ctx, solver)

	clockSignal := frameVars[assertion.ClockSignal][frame]

	antecedentAndClockSig := ctx.MkBVAnd(clockSignal, antecedent)
	antecedentTrue := ctx.MkEq(antecedentAndClockSig, trueBv)

	implication := ctx.MkImplies(antecedentTrue, consequent)

	assertionNot := ctx.MkNot(implication)
	fmt.Print(assertionNot.String())

	return assertionNot

}

func ExprToZ3(expr *nodes.ExprIR, frame int, frameVars map[string][]*z3.Expr, ctx *z3.Context, solver *z3.Solver) *z3.Expr {

	if expr == nil || expr.Kind == EmptyString {
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

		nestedExpr := ExprToZ3(expr.Left, frame, frameVars, ctx, solver)

		if nestedExpr != nil {

			diff := uint(expr.Width) - uint(expr.Left.Width)

			if diff > 0 {
				return ctx.MkZeroExt(diff, nestedExpr)
			}

			return nestedExpr

		}

		return nil

	case "BinaryOp":
		if expr.Op != EmptyString {

			if expr.Op == "LogicalAnd" {

				left := ExprToZ3(expr.Left, frame, frameVars, ctx, solver)
				right := ExprToZ3(expr.Right, frame, frameVars, ctx, solver)

				implication := ctx.MkBVAnd(left, right)

				// fmt.Print(left.String() + "\n")
				// fmt.Print(right.String() + "\n")
				// fmt.Print(implication.String() + "\n")

				return implication

			}

			if expr.Op == "Equality" {

				left := ExprToZ3(expr.Left, frame, frameVars, ctx, solver) //parsing empty val- ""
				right := ExprToZ3(expr.Right, frame, frameVars, ctx, solver)

				equality := ctx.MkEq(left, right) //this is giving the sorts
				// of bitvec 1 and 32 are incompatible

				return equality

			}

		}
	case "ConditionalOp":

		iteIf := ExprToZ3(expr.Args[0], frame, frameVars, ctx, solver)
		iteThen := ExprToZ3(expr.Args[1], frame, frameVars, ctx, solver)
		iteElse := ExprToZ3(expr.Args[2], frame, frameVars, ctx, solver)

		return ResolveConditionalITE(ctx, solver, expr, iteIf, iteThen, iteElse, frame)

	case "UnaryOp":

		if expr.Op != EmptyString {

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

func RecurseLeftsForSymbol(expr *nodes.ExprIR) string {

	if expr.Symbol != EmptyString {

		return expr.Symbol

	} else {

		if expr.Left != nil {

			return RecurseLeftsForSymbol(expr.Left)

		}

	}

	return EmptyString

}
