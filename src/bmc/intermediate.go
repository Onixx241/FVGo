package bmc

import (
	"log"
	"strconv"
	"strings"
)

func AstInstance(ast SlangAST) (dict map[string]*IRNode, design DesignGraph) {

	componentDict := make(map[string]*IRNode)
	var graph DesignGraph

	for _, current := range ast.Design.Members {

		if current.Kind == "Instance" {
			for _, component := range current.Body.Members {

				switch component.Kind {

				case "Port":
					componentDict[component.Name] = &IRNode{Name: component.Name, Type: InputType, Width: ParseBitWidth(component.Type)} //parse [lsb/msb:msb/lsb] later for bigger widths
					graph.Inputs = append(graph.Inputs, componentDict[component.Name])

				case "Variable":
					componentDict[component.Name] = &IRNode{Name: component.Name, Type: SignalType, Width: ParseBitWidth(component.Type)}
					graph.Signals = append(graph.Signals, componentDict[component.Name])

				case "ProceduralBlock":
					RecurseMember(component, &graph, componentDict)

				}

			}
		}

	}

	return componentDict, graph

}

func RecurseMember(member MemberNode, graph *DesignGraph, componentDict map[string]*IRNode) {
	switch member.Kind {
	case "ProceduralBlock":

		switch member.ProcedureKind {

		case "AlwaysFF":
			if member.Body != nil && member.Body.Statement != nil {
				RecurseStatement(*member.Body.Statement, nil, graph, member.ProcedureKind, componentDict)
			}

		case "AlwaysComb":
			panic("Not Implemented")

		}

	}
}

func RecurseStatement(stmt StatementNode, activeCondition *IRNode, graph *DesignGraph, ProcedureKind string, componentDict map[string]*IRNode) {
	switch stmt.Kind {

	case "Block":
		if stmt.Body != nil && stmt.Body.Statement != nil {
			RecurseStatement(*stmt.Body.Statement, activeCondition, graph, ProcedureKind, componentDict)
		}

	case "Conditional":
		returnedConditional := RecurseExpression(*stmt.Body.Conditions[0].Expr, nil, graph, componentDict, ProcedureKind)

		RecurseExpression(stmt.Body.IfTrue.Expr, returnedConditional, graph, componentDict, ProcedureKind)

	case "ExpressionStatement":
		RecurseExpression(*stmt.Expression, nil, graph, componentDict, ProcedureKind)

	}
}

func RecurseExpression(expr ExpressionNode, activeCondition *IRNode, graph *DesignGraph, componentDict map[string]*IRNode, ProcedureKind string) *IRNode {
	switch expr.Kind {

	case "NamedValue":
		return componentDict[ParseSymbol(*expr.Operand.Symbol)]

	case "UnaryOp":
		returnedExpr := RecurseExpression(ConvertOperand(expr.Operand), nil, graph, componentDict, ProcedureKind)

		GateNode := IRNode{Name: returnedExpr.Name, Type: GateType, Width: returnedExpr.Width}
		GateNode.Inputs = append(GateNode.Inputs, returnedExpr)

		return &GateNode
	case "Assignment":

		right := RecurseExpression(ConvertRightNode(expr.Right), nil, graph, componentDict, ProcedureKind)

		left := RecurseExpression(ConvertLeftNode(expr.Left), nil, graph, componentDict, ProcedureKind)

		componentLookup := componentDict[ParseSymbol(expr.Left.Symbol)]

		if ProcedureKind == "AlwaysFF" {

			if activeCondition != nil {

				newNode := IRNode{Name: "SignalComb", Type: GateType}
				newNode.Inputs = append(newNode.Inputs, activeCondition, right, componentLookup)
				componentLookup.NextState = &newNode

			} else {

				componentLookup.NextState = right

			}

		}
		if ProcedureKind == "AlwaysComb" {

			componentLookup.Inputs = append(componentLookup.Inputs, left)

		}

	case "Conversion":
		returnedConversionExpression := RecurseExpression(ConvertOperand(expr.Operand), nil, graph, componentDict, ProcedureKind)
		return returnedConversionExpression

	case "IntegerLiteral":
		val := expr.Constant
		return &IRNode{Type: ConstantType, Name: *val}
	}

	return nil
}

func ParseBitWidth(unparsed string) int {

	preparse := unparsed

	if unparsed == "logic" {
		return 1
	} else {

		preparse = strings.ReplaceAll(preparse, "[", "")
		preparse = strings.ReplaceAll(preparse, "logic", "")
		preparse = strings.ReplaceAll(preparse, "]", "")

		if strings.Contains(preparse, "signed") {
			preparse = strings.ReplaceAll(preparse, "signed", "")
		}

		nums := strings.Split(preparse, ":")

		leftnum, err := strconv.Atoi(nums[0])
		if err != nil {
			log.Fatal(err)
			return 0
		}
		rightnum, err := strconv.Atoi(nums[1])
		if err != nil {
			log.Fatal(err)
			return 0
		}

		finalnum := (leftnum - rightnum)

		if finalnum < 0 {
			finalnum--
			return -finalnum
		} else {
			finalnum++
			return finalnum
		}

	}
}

func ParseSymbol(name string) string {

	return strings.Split(name, " ")[1]

}

func ConvertOperand(expr *OperandNode) ExpressionNode {

	if expr.Operand.Symbol != nil {

		return ExpressionNode{Kind: expr.Operand.Kind, Type: expr.Operand.Type, Symbol: *expr.Operand.Symbol, Value: expr.Operand.Value, Constant: expr.Operand.Constant}

	} else {

		return ExpressionNode{Kind: expr.Operand.Kind, Type: expr.Operand.Type, Value: expr.Operand.Value, Constant: expr.Operand.Constant}

	}

}
func ConvertLeftNode(left *LeftNode) ExpressionNode {
	return ExpressionNode{Kind: left.Kind, Type: left.Type, Symbol: left.Symbol}
}
func ConvertRightNode(right *RightNode) ExpressionNode {
	return ExpressionNode{Kind: right.Kind, Type: right.Type, Operand: &right.Operand}
}
