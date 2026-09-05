package bmc

import (
	"strconv"
	"strings"
)

func AstInstance(ast SlangAST) (dict map[string]*IRNode, design DesignGraph) {
	componentDict := make(map[string]*IRNode)
	var graph DesignGraph

	for _, current := range ast.Design.Members {

		if current.Kind != "Instance" || current.Body == nil {
			continue

		}

		for _, component := range current.Body.Members {

			switch component.Kind {

			case "Port":
				addNode := &IRNode{Name: component.Name, Type: InputType, Width: ParseBitWidth(component.Type)}

				if component.Direction == "Out" {
					addNode.Type = OutputType
				}

				componentDict[component.Name] = addNode

				if addNode.Type == InputType {
					graph.Inputs = append(graph.Inputs, addNode)

				} else {
					graph.Outputs = append(graph.Outputs, addNode)

				}

			case "Variable":
				componentDict[component.Name] = &IRNode{Name: component.Name, Type: SignalType, Width: ParseBitWidth(component.Type)}
				graph.Signals = append(graph.Signals, componentDict[component.Name])

			case "Parameter":
				if component.Initializer != nil {

					paramNode := &ConstNode{

						Name:        component.Name,
						Kind:        component.Kind,
						Type:        LocalParamType,
						Width:       ParseBitWidth(component.Type),
						Initializer: *component.Initializer,
						Value:       component.Initializer.Constant,
					}

					graph.LocalParams = append(graph.LocalParams, paramNode)

				}

			case "ProceduralBlock":
				proc := &ProcessIR{Kind: component.ProcedureKind}

				if component.Body != nil && component.Body.Timing != nil {

					tempNode := &TimingNode{
						Kind:       component.Kind,
						Expression: component.Body.Timing.Expression,
						Edge:       component.Body.Timing.Edge,
					}

					graph.Processes = append(graph.Processes, tempNode)
					proc.Timing = tempNode

				}

				order := 0
				WalkBodyLower(component.Body, componentDict, &graph, proc, nil, &order)
				graph.ProcessIRs = append(graph.ProcessIRs, proc)

			}

		}
	}

	return componentDict, graph
}

func WalkBodyLower(body *BodyNode, dict map[string]*IRNode, graph *DesignGraph, proc *ProcessIR, guard *ExprIR, order *int) {

	if body == nil {
		return
	}

	if body.Kind == "ExpressionStatement" && body.Expression != nil {

		LowerAssignment(*body.Expression, dict, proc, guard, order)

	}

	if body.Kind == "Conditional" && body.Conditions != nil {

		trueGuard := guard

		for _, c := range body.Conditions {

			if c.Expr != nil {

				ex := LowerExpr(*c.Expr, dict)
				trueGuard = AndGuard(trueGuard, ex)

			}

		}

		if body.IfTrue != nil && body.IfTrue.Body != nil {

			WalkBodyLower(body.IfTrue.Body, dict, graph, proc, trueGuard, order)

		}

		if body.IfFalse != nil && body.IfFalse.Body != nil {

			falseGuard := guard

			for _, c := range body.Conditions {

				if c.Expr != nil {

					ex := LowerExpr(*c.Expr, dict)

					falseGuard = AndGuard(falseGuard, NotExpr(ex))

				}
			}

			WalkBodyLower(body.IfFalse.Body, dict, graph, proc, falseGuard, order)

		}
	}

	if body.List != nil {

		for _, listitem := range body.List {

			if listitem.Kind == "ExpressionStatement" && listitem.Expression != nil {

				LowerAssignment(*listitem.Expression, dict, proc, guard, order)

			}

			if listitem.Kind == "ProceduralAssign" && listitem.Assignment != nil {

				LowerAssignment(*listitem.Assignment, dict, proc, guard, order)

			}

			if listitem.Kind == "Conditional" && listitem.Conditions != nil {

				trueGuard := guard

				for _, c := range *listitem.Conditions {

					if c.Expr != nil {

						ex := LowerExpr(*c.Expr, dict)

						trueGuard = AndGuard(trueGuard, ex)

					}

				}

				if listitem.IfTrue != nil && listitem.IfTrue.Body != nil {

					WalkBodyLower(listitem.IfTrue.Body, dict, graph, proc, trueGuard, order)

				}

				if listitem.IfFalse != nil && listitem.IfFalse.Body != nil {

					falseGuard := guard

					for _, c := range *listitem.Conditions {

						if c.Expr != nil {

							ex := LowerExpr(*c.Expr, dict)

							falseGuard = AndGuard(falseGuard, NotExpr(ex))

						}

					}

					WalkBodyLower(listitem.IfFalse.Body, dict, graph, proc, falseGuard, order)

				}

			}
		}
	} else {

		WalkStatementLower(body.Statement, dict, graph, proc, guard, order)

	}

	if body.Kind == "Case" && body.Expression != nil && body.Items != nil {

		selector := LowerExpr(*body.Expression, dict)

		for _, item := range body.Items {

			itemGuard := guard

			if item.Expressions != nil {

				var orExpr *ExprIR

				for _, caseExpr := range item.Expressions {

					eq := &ExprIR{

						Kind:  "BinaryOp",
						Op:    "Equality",
						Left:  selector,
						Right: LowerExpr(*caseExpr, dict),
					}

					if orExpr == nil {

						orExpr = eq

					} else {

						orExpr = &ExprIR{Kind: "BinaryOp", Op: "LogicalOr", Left: orExpr, Right: eq}

					}

				}

				itemGuard = AndGuard(itemGuard, orExpr)

			}

			if item.Statement != nil {

				WalkStatementLower(item.Statement, dict, graph, proc, itemGuard, order)

			}

		}
	}
}

func WalkStatementLower(stmt *StatementNode, dict map[string]*IRNode, graph *DesignGraph, proc *ProcessIR, guard *ExprIR, order *int) {

	if stmt == nil {
		return
	}

	if stmt.Expression != nil {

		LowerAssignment(*stmt.Expression, dict, proc, guard, order)

	}

	if stmt.Body != nil {

		WalkBodyLower(stmt.Body, dict, graph, proc, guard, order)

	}

}

func LowerAssignment(node ExpressionNode, dict map[string]*IRNode, proc *ProcessIR, guard *ExprIR, order *int) {

	if node.Left == nil || node.Right == nil {
		return
	}

	targetName := ParseSymbol(node.Left.Symbol)

	target := dict[targetName]

	if target == nil {

		target = &IRNode{Name: targetName, Type: SignalType, Width: ParseBitWidth(node.Left.Type)}

		dict[targetName] = target

	}

	value := LowerRight(*node.Right, dict)

	nb := false

	if node.NonBlockingBool != nil {
		nb = *node.NonBlockingBool
	}

	assign := &GuardedAssign{

		Target:      target,
		Value:       value,
		Guard:       guard,
		NonBlocking: nb,
		ProcessKind: proc.Kind,
		Order:       *order,
	}

	*order++

	proc.Assignments = append(proc.Assignments, assign)

	signalComb := IRNode{Name: "SignalComb", Type: SignalType}

	target.NextState = &signalComb

}

func LowerRight(node RightNode, dict map[string]*IRNode) *ExprIR {

	if node.Symbol != "" {

		return SymbolExpr(node.Symbol, node.Type)

	}

	if node.Constant != nil {

		return &ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: *node.Constant, Width: ParseBitWidth(node.Type)}

	}

	return LowerOperand(&node.Operand, dict)
}

func LowerExpr(node ExpressionNode, dict map[string]*IRNode) *ExprIR {

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

		return &ExprIR{Kind: "IntegerLiteral", Type: node.Type, Value: v, Width: ParseBitWidth(node.Type)}

	case "Conversion":

		if node.Operand != nil {

			return &ExprIR{

				Kind:  "Conversion",
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerOperand(node.Operand, dict),
			}

		}

		return &ExprIR{Kind: "Conversion", Type: node.Type, Width: ParseBitWidth(node.Type)}

	case "UnaryOp":

		if node.Operand != nil {

			return &ExprIR{

				Kind:  "UnaryOp",
				Op:    node.Op,
				Type:  node.Type,
				Width: ParseBitWidth(node.Type),
				Left:  LowerOperand(node.Operand, dict),
			}

		}

		return &ExprIR{Kind: "UnaryOp", Op: node.Op, Type: node.Type, Width: ParseBitWidth(node.Type)}

	case "BinaryOp":

		e := &ExprIR{Kind: "BinaryOp", Op: node.Op, Type: node.Type, Width: ParseBitWidth(node.Type)}

		if node.Left != nil {
			e.Left = LowerExpr(ConvertLeftNode(node.Left), dict)
		}
		if node.Right != nil {
			e.Right = LowerRight(*node.Right, dict)
		}

		return e

	case "ConditionalOp":

		c := &ExprIR{Kind: "ConditionalOp", Type: node.Type, Width: ParseBitWidth(node.Type)}

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

	if node.Operand != nil {

		return LowerOperand(node.Operand, dict)

	}

	if node.Symbol != "" {

		return SymbolExpr(node.Symbol, node.Type)

	}

	return &ExprIR{Kind: node.Kind, Type: node.Type, Width: ParseBitWidth(node.Type)}
}

func LowerOperand(op *OperandNode, dict map[string]*IRNode) *ExprIR {

	if op == nil {
		return &ExprIR{Kind: "Nil"}
	}

	switch op.Kind {

	case "NamedValue":

		if op.Symbol != nil {
			return SymbolExpr(*op.Symbol, op.Type)
		}

		return &ExprIR{Kind: "NamedValue", Type: op.Type, Width: ParseBitWidth(op.Type)}

	case "IntegerLiteral":

		v := ""

		if op.Constant != nil {
			v = *op.Constant
		} else if op.Value != nil {
			v = *op.Value
		}

		return &ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: v, Width: ParseBitWidth(op.Type)}

	case "Conversion":

		return &ExprIR{

			Kind:  "Conversion",
			Type:  op.Type,
			Width: ParseBitWidth(op.Type),
			Left:  LowerOperand(op.Operand, dict),
		}

	case "UnaryOp":

		return &ExprIR{

			Kind:  "UnaryOp",
			Op:    op.Op,
			Type:  op.Type,
			Width: ParseBitWidth(op.Type),
			Left:  LowerOperand(op.Operand, dict),
		}

	case "BinaryOp":

		e := &ExprIR{Kind: "BinaryOp", Op: op.Op, Type: op.Type, Width: ParseBitWidth(op.Type)}

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

		return &ExprIR{Kind: "IntegerLiteral", Type: op.Type, Value: *op.Constant, Width: ParseBitWidth(op.Type)}

	}

	if op.Operand != nil {

		return LowerOperand(op.Operand, dict)

	}

	return &ExprIR{Kind: op.Kind, Type: op.Type, Width: ParseBitWidth(op.Type)}
}

func SymbolExpr(raw string, typ string) *ExprIR {

	return &ExprIR{

		Kind:   "NamedValue",
		Type:   typ,
		Width:  ParseBitWidth(typ),
		Symbol: ParseSymbol(raw),
	}

}

func AndGuard(a *ExprIR, b *ExprIR) *ExprIR {

	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	return &ExprIR{Kind: "BinaryOp", Op: "LogicalAnd", Left: a, Right: b, Type: "logic", Width: 1}

}

func NotExpr(a *ExprIR) *ExprIR {

	if a == nil {
		return nil
	}

	return &ExprIR{Kind: "UnaryOp", Op: "LogicalNot", Left: a, Type: "logic", Width: 1}

}

func ParseLiteralValue(unparsed string) int {

	panic("not implemented")

}

func ParseBitWidth(unparsed string) int {

	s := strings.TrimSpace(unparsed)

	if s == "" {
		return 1
	}

	if strings.HasPrefix(s, "logic") || strings.HasPrefix(s, "bit") {

		l := strings.Index(s, "[")

		r := strings.Index(s, "]")

		if l >= 0 && r > l {

			rangePart := s[l+1 : r]

			nums := strings.Split(rangePart, ":")

			if len(nums) == 2 {
				leftnum, err := strconv.Atoi(strings.TrimSpace(nums[0]))

				if err == nil {

					rightnum, err2 := strconv.Atoi(strings.TrimSpace(nums[1]))

					if err2 == nil {

						if leftnum >= rightnum {
							return leftnum - rightnum + 1

						}

						return rightnum - leftnum + 1
					}
				}
			}
		}
		return 1
	}

	if strings.Contains(s, "[") && strings.Contains(s, "]") {

		l := strings.Index(s, "[")

		r := strings.Index(s, "]")

		rangePart := s[l+1 : r]

		nums := strings.Split(rangePart, ":")

		if len(nums) == 2 {

			leftnum, err := strconv.Atoi(strings.TrimSpace(nums[0]))

			if err == nil {

				rightnum, err2 := strconv.Atoi(strings.TrimSpace(nums[1]))

				if err2 == nil {

					if leftnum >= rightnum {

						return leftnum - rightnum + 1

					}

					return rightnum - leftnum + 1
				}
			}
		}
	}

	return 32
}

func ParseSymbol(name string) string {

	parts := strings.Fields(name)

	if len(parts) == 0 {

		return name

	}

	return parts[len(parts)-1]
}

func ConvertOperand(operand *OperandNode) ExpressionNode {

	if operand == nil {

		return ExpressionNode{}

	}

	if operand.Operand == nil {

		sym := ""

		if operand.Symbol != nil {

			sym = *operand.Symbol

		}

		return ExpressionNode{

			Kind:     operand.Kind,
			Type:     operand.Type,
			Symbol:   sym,
			Value:    operand.Value,
			Constant: operand.Constant,
		}

	}
	return ExpressionNode{

		Kind:     operand.Operand.Kind,
		Type:     operand.Operand.Type,
		Value:    operand.Operand.Value,
		Constant: operand.Operand.Constant,
		Operand:  operand.Operand.Operand,
		Op:       operand.Operand.Op,
		Left:     nil,
		Right:    nil,
	}

}

func OperandSelectorHelper(ExpressionOperand *OperandNode, Selector string) *string {

	if ExpressionOperand == nil {

		empty := ""
		return &empty

	}

	switch Selector {

	case "Value":

		if ExpressionOperand.Operand != nil && ExpressionOperand.Operand.Value != nil {
			return ExpressionOperand.Operand.Value
		}

	case "Constant":

		if ExpressionOperand.Operand != nil && ExpressionOperand.Operand.Constant != nil {
			return ExpressionOperand.Operand.Constant
		}

	}

	empty := ""

	return &empty

}

func ConvertLeftNode(left *LeftNode) ExpressionNode {
	return ExpressionNode{Kind: left.Kind, Type: left.Type, Symbol: left.Symbol}
}

func ConvertRightNode(right *RightNode) ExpressionNode {
	return ExpressionNode{Kind: right.Kind, Type: right.Type, Operand: &right.Operand, Constant: right.Constant, Symbol: right.Symbol}
}

func ParseGateType(gateString string) LogicalType {

	switch gateString {

	case "LogicalNot":
		return NotGate
	case "LogicalAnd":
		return AndGate
	case "LogicalOr":
		return OrGate
	case "LogicalShiftLeft":
		return ShiftLeftGate
	default:
		return NilGate

	}
}

func ProcessExpressionStatement(node ExpressionNode, dict map[string]*IRNode) IRNode {

	signalComb := &IRNode{Name: "SignalComb", Type: SignalType}

	if node.Right != nil {

		if node.Right.Symbol != "" {

			ref := ParseSymbol(node.Right.Symbol)

			if dict[ref] != nil {

				signalComb.Inputs = append(signalComb.Inputs, dict[ref])

			}

		} else if node.Right.Constant != nil {

			c := &IRNode{Name: *node.Right.Constant, Type: ConstantType, Width: ParseBitWidth(node.Right.Type)}

			signalComb.Inputs = append(signalComb.Inputs, c)

		} else if node.Right.Operand.Op != "" {

			signalComb.Op = node.Right.Operand.Op

		}

	}
	return *signalComb
}
