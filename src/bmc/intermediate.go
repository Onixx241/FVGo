package bmc

import (
	"log"
	"main/slang"
	"strconv"
	"strings"
)

type ComponentType int

const (
	InputType ComponentType = iota
	OutputType
	StateType
	SignalType
)

type StatementNode struct {
	Kind       string
	BlockKind  string
	Conditions ConditionalNode
}
type ConditionalNode struct {
	Kind    string
	Type    string
	Op      string
	Operand OperandNode
}
type OperandNode struct {
	Kind     string
	Type     string
	Symbol   *string
	Operand  *OperandNode
	Value    *string
	Constant *string
}

type TruthNode struct {
	Kind       string
	Expression ExpressionNode
}

type ExpressionNode struct {
	Kind            string
	Type            string
	Left            LeftNode
	Right           RightNode
	NonBlockingBool bool
}
type LeftNode struct {
	Kind   string
	Type   string
	Symbol string
}
type RightNode struct {
	Kind    string
	Type    string
	Operand OperandNode
}

type IRNode struct {
	Name  string
	Type  ComponentType
	Width int
}

type DesignGraph struct {
	Inputs  []IRNode
	States  []IRNode
	Signals []IRNode
}

func AstInstance(ast slang.SlangAST) map[string]IRNode {

	componentDict := make(map[string]IRNode)
	var graph DesignGraph

	for _, current := range ast.Design.Members {

		if current.Kind == "Instance" {
			for _, component := range current.Body.Members {

				switch component.Kind {
				case "Port":
					componentDict[component.Name] = IRNode{Name: component.Name, Type: InputType, Width: ParseBitWidth(component.Type)} //parse [lsb/msb:msb/lsb] later for bigger widths
					graph.Inputs = append(graph.Inputs, IRNode{Name: component.Name, Type: InputType, Width: ParseBitWidth(component.Type)})
				case "Variable":
					componentDict[component.Name] = IRNode{Name: component.Name, Type: SignalType, Width: ParseBitWidth(component.Type)}
					graph.Signals = append(graph.Signals, IRNode{Name: component.Name, Type: SignalType, Width: ParseBitWidth(component.Type)})

				}

			}
		}

	}

	return componentDict

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

		finalnum := leftnum - rightnum

		if rightnum == 0 || leftnum == 0 {

			if finalnum < 0 {

				finalnum--

			} else {

				finalnum++

			}

		}

		if finalnum < 0 {
			return -finalnum
		} else {
			return finalnum
		}

	}
}

func ParseSymbol(name string) { //return string when i implement
	panic("not implemented")
}

func RecurseAST(member slang.MemberNode, graph *DesignGraph) {
	panic("not implemented")
}
