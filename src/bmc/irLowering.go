package bmc

import (
	"main/bmc/nodes"
)

func LowerAssignment(node nodes.ExpressionNode, dict map[string]*nodes.IRNode, proc *nodes.ProcessIR, guard *nodes.ExprIR, order *int) {

	if node.Left == nil || node.Right == nil {
		return
	}

	targetName := ParseSymbol(node.Left.Symbol)

	target := dict[targetName]

	if target == nil {

		target = &nodes.IRNode{Name: targetName, Type: nodes.SignalType, Width: ParseBitWidth(node.Left.Type)}

		dict[targetName] = target

	}

	value := LowerRight(*node.Right, dict)

	nb := false

	if node.NonBlockingBool != nil {
		nb = *node.NonBlockingBool
	}

	assign := &nodes.GuardedAssign{

		Target:      target,
		Value:       value,
		Guard:       guard,
		NonBlocking: nb,
		ProcessKind: proc.Kind,
		Order:       *order,
	}

	*order = *order + 1

	proc.Assignments = append(proc.Assignments, assign)

	signalComb := nodes.IRNode{Name: "SignalComb", Type: nodes.SignalType}

	target.NextState = &signalComb

}

func LowerRight(node nodes.RightNode, dict map[string]*nodes.IRNode) *nodes.ExprIR {

	if node.Symbol != "" {
		return SymbolExpr(node.Symbol, node.Type)
	}

	if node.Constant != nil {
		return &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: *node.Constant, Width: ParseBitWidth(node.Type)}
	}

	if node.Kind == "ConditionalOp" {

		c := &nodes.ExprIR{Kind: "ConditionalOp", Type: node.Type, Width: ParseBitWidth(node.Type)}

		if node.Conditions != nil && len(*node.Conditions) > 0 {
			if (*node.Conditions)[0].Expr != nil {
				c.Args = append(c.Args, LowerExpr(*(*node.Conditions)[0].Expr, dict))
			}
		}

		if node.Left != nil {
			c.Args = append(c.Args, LowerExpr(ConvertLeftNode(node.Left), dict))
		}

		if node.Right != nil {
			c.Args = append(c.Args, LowerRight(*node.Right, dict))
		}

		return c
	}

	if node.Kind == "BinaryOp" {

		e := &nodes.ExprIR{Kind: "BinaryOp", Op: node.Op, Type: node.Type, Width: ParseBitWidth(node.Type)}

		if node.Left != nil {
			e.Left = LowerExpr(ConvertLeftNode(node.Left), dict)
		}

		if node.Right != nil {
			e.Right = LowerRight(*node.Right, dict)
		}

		return e
	}

	if node.Kind == "Conversion" {

		if node.Symbol != "" {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  SymbolExpr(node.Symbol, node.Type),
			}
		}

		if node.Constant != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: *node.Constant, Width: ParseBitWidth(node.Type)},
			}
		}

		opExpr := LowerOperand(&node.Operand, dict)

		if opExpr == nil || opExpr.Kind == "" || opExpr.Kind == "Nil" {

			opExpr = &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: "0", Width: ParseBitWidth(node.Type)}
		}

		return &nodes.ExprIR{
			Kind:  "Conversion",
			Type:  node.Type,
			Width: ParseBitWidth(node.Type),
			Left:  opExpr,
		}
	}

	return LowerOperand(&node.Operand, dict)
}

func LowerExpr(node nodes.ExpressionNode, dict map[string]*nodes.IRNode) *nodes.ExprIR {

	switch node.Kind {

	case "NamedValue":
		return SymbolExpr(node.Symbol, node.Type)

	case "IntegerLiteral":

		v := ""

		if node.Constant != nil {
			v = *node.Constant
		} else if node.Value != nil {
			v = *node.Value
		}

		return &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: v, Width: ParseBitWidth(node.Type)}

	case "Conversion":

		if node.Operand != nil {

			return &nodes.ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerOperand(node.Operand, dict),
			}

		}

		if node.Left != nil {

			return &nodes.ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerExpr(ConvertLeftNode(node.Left), dict),
			}

		}

		if node.Symbol != "" {

			return &nodes.ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  SymbolExpr(node.Symbol, node.Type),
			}

		}

		if node.Constant != nil {

			return &nodes.ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: *node.Constant, Width: ParseBitWidth(node.Type)},
			}

		}

		if node.Value != nil {

			return &nodes.ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  &nodes.ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: *node.Value, Width: ParseBitWidth(node.Type)},
			}

		}

		if node.OperandExpr != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerExpr(*node.OperandExpr, dict),
			}
		}

		return &nodes.ExprIR{Kind: "Conversion", Type: node.Type, Width: ParseBitWidth(node.Type), Left: &nodes.ExprIR{Kind: "Nil"}}

	case "UnaryOp":

		if node.Operand != nil {

			return &nodes.ExprIR{

				Kind:  "UnaryOp",
				Op:    node.Op,
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerOperand(node.Operand, dict),
			}

		}

		return &nodes.ExprIR{Kind: "UnaryOp", Op: node.Op, Type: node.Type, Width: ParseBitWidth(node.Type)}

	case "BinaryOp":

		e := &nodes.ExprIR{Kind: "BinaryOp", Op: node.Op, Type: node.Type, Width: ParseBitWidth(node.Type)}

		if node.Left != nil {
			e.Left = LowerExpr(ConvertLeftNode(node.Left), dict)
		}
		if node.Right != nil {
			e.Right = LowerRight(*node.Right, dict)
		}

		return e

	case "ConditionalOp":

		c := &nodes.ExprIR{Kind: "ConditionalOp", Type: node.Type, Width: ParseBitWidth(node.Type)}

		if node.Conditions != nil && len(*node.Conditions) > 0 {

			if (*node.Conditions)[0].Expr != nil {
				c.Args = append(c.Args, LowerExpr(*(*node.Conditions)[0].Expr, dict))
			}

		}

		if node.Left != nil {

			c.Args = append(c.Args, LowerExpr(ConvertLeftNode(node.Left), dict))

		}

		if node.Right != nil {

			c.Args = append(c.Args, LowerRight(*node.Right, dict))

		}

		return c

	case "Simple": //assumption property - continue here

		assumption := &nodes.ExprIR{Kind: "AssumptionProperty"}

		innerExpr := LowerExpr(*node.Expression, dict) //continue here

		if innerExpr.Op == "Equality" {

			assumption.Op = innerExpr.Op
			assumption.Value = innerExpr.Right.Value
			assumption.Width = innerExpr.Right.Width
			constraint := &nodes.ExprIR{Width: innerExpr.Left.Width, Symbol: innerExpr.Left.Symbol}
			assumption.Args = append(assumption.Args, constraint)

		} else {

			assumption.Type = innerExpr.Type
			assumption.Op = innerExpr.Op
			assumption.Width = innerExpr.Width

			constraint := &nodes.ExprIR{Width: innerExpr.Left.Width, Symbol: innerExpr.Left.Symbol}
			assumption.Args = append(assumption.Args, constraint)

		}

		return assumption

	}

	if node.Operand != nil {

		return LowerOperand(node.Operand, dict)

	}

	if node.Symbol != "" {

		return SymbolExpr(node.Symbol, node.Type)

	}

	return &nodes.ExprIR{Kind: node.Kind, Type: node.Type, Width: ParseBitWidth(node.Type)}
}

func LowerOperand(op *nodes.OperandNode, dict map[string]*nodes.IRNode) *nodes.ExprIR {

	if op == nil {
		return &nodes.ExprIR{Kind: "Nil"}
	}

	switch op.Kind {

	case "NamedValue":

		if op.Symbol != nil {
			return SymbolExpr(*op.Symbol, op.Type)
		}

		return &nodes.ExprIR{Kind: "NamedValue", Type: op.Type, Width: ParseBitWidth(op.Type)}

	case "IntegerLiteral":

		v := ""

		if op.Constant != nil {
			v = *op.Constant
		} else if op.Value != nil {
			v = *op.Value
		}

		return &nodes.ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: v, Width: ParseBitWidth(op.Type)}

	case "Conversion":

		if op.Operand != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  op.Type,
				Width: ParseBitWidth(op.Type),
				Left:  LowerOperand(op.Operand, dict),
			}
		}

		if op.Symbol != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  op.Type,
				Width: ParseBitWidth(op.Type),
				Left:  SymbolExpr(*op.Symbol, op.Type),
			}
		}

		if op.Constant != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  op.Type,
				Width: ParseBitWidth(op.Type),
				Left:  &nodes.ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: *op.Constant, Width: ParseBitWidth(op.Type)},
			}
		}

		if op.Value != nil {
			return &nodes.ExprIR{
				Kind:  "Conversion",
				Type:  op.Type,
				Width: ParseBitWidth(op.Type),
				Left:  &nodes.ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: *op.Value, Width: ParseBitWidth(op.Type)},
			}
		}

		return &nodes.ExprIR{
			Kind:  "Conversion",
			Type:  op.Type,
			Width: ParseBitWidth(op.Type),
			Left:  &nodes.ExprIR{Kind: "Nil"},
		}

	case "UnaryOp":

		return &nodes.ExprIR{
			Kind:  "UnaryOp",
			Op:    op.Op,
			Type:  op.Type,
			Width: ParseBitWidth(op.Type),
			Left:  LowerOperand(op.Operand, dict),
		}

	case "BinaryOp":

		e := &nodes.ExprIR{Kind: "BinaryOp", Op: op.Op, Type: op.Type, Width: ParseBitWidth(op.Type)}

		if op.Left != nil {
			e.Left = LowerExpr(*op.Left, dict)
		}

		if op.Right != nil {
			e.Right = LowerExpr(*op.Right, dict)
		}

		return e
	}

	if op.Symbol != nil {
		return SymbolExpr(*op.Symbol, op.Type)
	}

	if op.Constant != nil {
		return &nodes.ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: *op.Constant, Width: ParseBitWidth(op.Type)}
	}

	if op.Value != nil {
		return &nodes.ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: *op.Value, Width: ParseBitWidth(op.Type)}
	}

	if op.Operand != nil {
		return LowerOperand(op.Operand, dict)
	}

	return &nodes.ExprIR{Kind: op.Kind, Type: op.Type, Width: ParseBitWidth(op.Type)}
}

func LowerConcurrentAssertion(body *nodes.BodyNode, dict map[string]*nodes.IRNode) *nodes.AssertionIR {

	if body == nil || body.PropertySpec == nil || body.PropertySpec.Expr == nil {
		return nil
	}

	property := body.PropertySpec.Expr

	if property.Kind != "Binary" { //fixed crash
		return nil
	}

	out := &nodes.AssertionIR{

		Kind:        body.AssertionKind,
		Implication: property.Op,
	}

	if body.PropertySpec.Clocking != nil {
		out.ClockEdge = body.PropertySpec.Clocking.Edge
		out.ClockSignal = ParseSymbol(body.PropertySpec.Clocking.Expression.Symbol)
	}
	//take into account sensitivity list with i rst or other async signals
	if property.Left != nil && property.Left.Expr != nil {
		out.Antecedent = LowerExpr(*property.Left.Expr, dict)
	}

	if property.Right != nil {

		if property.Right.Kind == "SequenceConcat" && len(property.Right.Elements) > 0 {
			element := property.Right.Elements[0]

			out.DelayMin = element.Min
			out.DelayMax = element.Max

			if element.Sequence != nil && element.Sequence.Expr != nil {

				out.Consequent = LowerExpr(*element.Sequence.Expr, dict)

			}

		} else if property.Right.Expr != nil {

			out.DelayMin = 0
			out.DelayMax = 0
			out.Consequent = LowerExpr(*property.Right.Expr, dict)
		}

	}

	return out

}

func LowerAssumeProperty(body *nodes.BodyNode, dict map[string]*nodes.IRNode) *nodes.AssumptionIR {

	if body == nil || body.PropertySpec == nil {
		return nil
	}

	newAssumption := nodes.AssumptionIR{Kind: body.PropertySpec.Kind}
	newAssumption.ClockEdge = body.PropertySpec.Expr.Body.Clocking.Edge
	newAssumption.ClockSignal = ParseSymbol(body.PropertySpec.Expr.Body.Clocking.Expression.Symbol)

	expr := LowerExpr(*body.PropertySpec.Expr.Body.Expression, dict)

	newAssumption.AssumptionExpr = expr
	newAssumption.ConstraintSymbol = expr.Args[0].Symbol

	return &newAssumption

}
