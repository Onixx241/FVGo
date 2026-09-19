package bmc

import (
	"log"
	"main/bmc/nodes"
	"strconv"
	"strings"
)

func ExtractAssignmentFromListItem(listitem *nodes.ListNode) *nodes.ExpressionNode {

	if listitem == nil {
		return nil
	}

	if listitem.Kind == "ProceduralAssign" && listitem.Assignment != nil && listitem.Assignment.Kind == "Assignment" {
		return listitem.Assignment
	}

	if listitem.Kind == "ExpressionStatement" && listitem.Expression != nil && listitem.Expression.Kind == "Assignment" {
		return listitem.Expression
	}

	if listitem.Assignment != nil && listitem.Assignment.Kind == "Assignment" {
		return listitem.Assignment
	}

	if listitem.Expression != nil && listitem.Expression.Kind == "Assignment" {
		return listitem.Expression
	}

	return nil
}

func ParseLiteralValue(unparsed string) int {

	if unparsed == "logic" {
		return 1
	}

	split := strings.Split(unparsed, "'")

	if len(split) == 1 {

		value, err := strconv.Atoi(unparsed)

		if err != nil {
			log.Fatal(err)
		}

		return value

	}

	width, err := strconv.Atoi(split[0])

	if err != nil {
		log.Fatal(err)
	}

	radix := string(split[1][0])
	valStr := split[1][1:]

	var base int
	switch radix {

	case "b", "B":
		base = 2

	case "h", "H":
		base = 16

	case "d", "D":
		base = 10

	case "o", "O":
		base = 8

	default:
		return 0

	}

	value, err := strconv.ParseInt(valStr, base, 64)

	if err != nil {
		log.Fatal(err)
	}

	_ = width

	return int(value)

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

func ConvertOperand(operand *nodes.OperandNode) nodes.ExpressionNode {

	if operand == nil {

		return nodes.ExpressionNode{}

	}

	if operand.Operand == nil {

		sym := ""

		if operand.Symbol != nil {

			sym = *operand.Symbol

		}

		return nodes.ExpressionNode{

			Kind:     operand.Kind,
			Type:     operand.Type,
			Symbol:   sym,
			Value:    operand.Value,
			Constant: operand.Constant,
		}

	}
	return nodes.ExpressionNode{

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

func OperandSelectorHelper(ExpressionOperand *nodes.OperandNode, Selector string) *string {

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

func ConvertLeftNode(left *nodes.LeftNode) nodes.ExpressionNode {

	if left == nil {
		return nodes.ExpressionNode{}
	}

	return nodes.ExpressionNode{
		Kind:       left.Kind,
		Type:       left.Type,
		Symbol:     left.Symbol,
		Operand:    left.Operand,
		Left:       left.Left,
		Right:      left.Right,
		Op:         left.Op,
		Conditions: left.Conditions,
		Value:      left.Value,
		Constant:   left.Constant,
	}
}

func ConvertRightNode(right *nodes.RightNode) nodes.ExpressionNode {
	return nodes.ExpressionNode{Kind: right.Kind, Type: right.Type, Operand: &right.Operand, Constant: right.Constant, Symbol: right.Symbol}
}

func ParseGateType(gateString string) nodes.LogicalType {

	switch gateString {

	case "LogicalNot":
		return nodes.NotGate
	case "LogicalAnd":
		return nodes.AndGate
	case "LogicalOr":
		return nodes.OrGate
	case "LogicalShiftLeft":
		return nodes.ShiftLeftGate
	default:
		return nodes.NilGate

	}
}
