package bmc

import (
	"log"
	"main/slang"
	"strconv"
	"strings"
)

func AstInstance(ast slang.SlangAST) (dict map[string]IRNode, design DesignGraph) {

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

	return componentDict, graph

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

func RecurseAST(member slang.MemberNode, graph *DesignGraph) {
	panic("not implemented")
	// switch member.Kind {
	// case "ProceduralBlock":
	// 	RecurseAST(member.Body.)

	// }
}
