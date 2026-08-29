package bmc

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
