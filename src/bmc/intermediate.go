package bmc

import (
	"main/bmc/nodes"
)

func AstInstance(ast nodes.SlangAST) (dict map[string]*nodes.IRNode, design nodes.DesignGraph) {
	componentDict := make(map[string]*nodes.IRNode)
	var graph nodes.DesignGraph

	for _, current := range ast.Design.Members {

		if current.Kind != "Instance" || current.Body == nil {
			continue

		}

		for _, component := range current.Body.Members {

			switch component.Kind {

			case "Port":
				addNode := &nodes.IRNode{Name: component.Name, Type: nodes.InputType, Width: ParseBitWidth(component.Type)}

				if component.Direction == "Out" {
					addNode.Type = nodes.OutputType
				}

				componentDict[component.Name] = addNode

				if addNode.Type == nodes.InputType {
					graph.Inputs = append(graph.Inputs, addNode)

				} else {
					graph.Outputs = append(graph.Outputs, addNode)

				}

			case "Variable":
				componentDict[component.Name] = &nodes.IRNode{Name: component.Name, Type: nodes.SignalType, Width: ParseBitWidth(component.Type)}
				graph.Signals = append(graph.Signals, componentDict[component.Name])

			case "Parameter":
				if component.Initializer != nil {

					paramNode := &nodes.ConstNode{

						Name:        component.Name,
						Kind:        component.Kind,
						Type:        nodes.LocalParamType,
						Width:       ParseBitWidth(component.Type),
						Initializer: *component.Initializer,
						Value:       component.Initializer.Constant,
					}

					graph.LocalParams = append(graph.LocalParams, paramNode)

				}

			case "ProceduralBlock":

				if component.Body != nil && component.Body.Kind == "ConcurrentAssertion" && component.Body.PropertySpec != nil {

					a := LowerConcurrentAssertion(component.Body, componentDict)

					if a != nil {
						graph.AssertionIRs = append(graph.AssertionIRs, a)
					}

				}

				proc := &nodes.ProcessIR{Kind: component.ProcedureKind}

				if component.Body != nil && component.Body.Timing != nil {

					tempNode := &nodes.TimingNode{
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

func WalkBodyLower(body *nodes.BodyNode, dict map[string]*nodes.IRNode, graph *nodes.DesignGraph, proc *nodes.ProcessIR, guard *nodes.ExprIR, order *int) {

	if body == nil {
		return
	}

	if body.Body != nil {
		WalkBodyLower(body.Body, dict, graph, proc, guard, order)
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

			if listitem == nil {
				continue
			}

			expr := ExtractAssignmentFromListItem(listitem)

			if expr != nil {
				LowerAssignment(*expr, dict, proc, guard, order)
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

				var orExpr *nodes.ExprIR

				for _, caseExpr := range item.Expressions {

					eq := &nodes.ExprIR{

						Kind:  "BinaryOp",
						Op:    "Equality",
						Left:  selector,
						Right: LowerExpr(*caseExpr, dict),
					}

					if orExpr == nil {

						orExpr = eq

					} else {

						orExpr = &nodes.ExprIR{Kind: "BinaryOp", Op: "LogicalOr", Left: orExpr, Right: eq}

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

func WalkStatementLower(stmt *nodes.StatementNode, dict map[string]*nodes.IRNode, graph *nodes.DesignGraph, proc *nodes.ProcessIR, guard *nodes.ExprIR, order *int) {

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

func SymbolExpr(raw string, typ string) *nodes.ExprIR {

	return &nodes.ExprIR{

		Kind:   "NamedValue",
		Type:   typ,
		Width:  ParseBitWidth(typ),
		Symbol: ParseSymbol(raw),
	}

}

func AndGuard(a *nodes.ExprIR, b *nodes.ExprIR) *nodes.ExprIR {

	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	return &nodes.ExprIR{Kind: "BinaryOp", Op: "LogicalAnd", Left: a, Right: b, Type: "logic", Width: 1}

}

func NotExpr(a *nodes.ExprIR) *nodes.ExprIR {

	if a == nil {
		return nil
	}

	return &nodes.ExprIR{Kind: "UnaryOp", Op: "LogicalNot", Left: a, Type: "logic", Width: 1}

}
