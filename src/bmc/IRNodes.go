package bmc

type ComponentType int

const (
	InputType ComponentType = iota
	OutputType
	StateType
	SignalType
	GateType
	ConstantType
)

type TimingNode struct {
	Kind       string         `json:"kind"`
	Expression ExpressionNode `json:"expr"`
	Edge       string         `json:"edge"`
}

type StatementNode struct {
	Kind       string            `json:"kind"`
	BlockKind  string            `json:"blockKind"`
	Conditions []ConditionalNode `json:"conditions"`
	Body       *BodyNode         `json:"body"`
	Expression *ExpressionNode   `json:"expr"`
}
type ConditionalNode struct {
	Kind    string          `json:"kind"`
	Type    string          `json:"type"`
	Op      string          `json:"op"`
	Operand OperandNode     `json:"operand"`
	Expr    *ExpressionNode `json:"expr"`
}
type OperandNode struct {
	Kind     string       `json:"kind"`
	Type     string       `json:"type"`
	Symbol   *string      `json:"symbol"`
	Operand  *OperandNode `json:"operand"`
	Value    *string      `json:"value"`
	Constant *string      `json:"constant"`
}

type TruthNode struct {
	Kind string         `json:"kind"`
	Expr ExpressionNode `json:"expr"`
}

type ExpressionNode struct { //acts like a general node sometimes, maybe make a dedicated general node?
	Kind            string       `json:"kind"`
	Type            string       `json:"type"`
	Symbol          string       `json:"symbol"`
	Left            *LeftNode    `json:"left,omitempty"`
	Right           *RightNode   `json:"right,omitempty"`
	Op              string       `json:"op,omitempty"`
	Operand         *OperandNode `json:"operand,omitempty"`
	NonBlockingBool *bool        `json:"isNonBlocking,omitempty"`
	Value           *string      `json:"value,omitempty"`
	Constant        *string      `json:"constant,omitempty"`
}
type LeftNode struct {
	Kind   string `json:"kind"`
	Type   string `json:"type"`
	Symbol string `json:"symbol"`
}
type RightNode struct {
	Kind     string      `json:"kind"`
	Type     string      `json:"type"`
	Operand  OperandNode `json:"operand"`
	Constant *string     `json:"constant,omitempty"`
}

type IRNode struct {
	Name      string
	Type      ComponentType
	Width     int
	Inputs    []*IRNode
	NextState *IRNode
}

type DesignGraph struct {
	Inputs  []*IRNode
	States  []*IRNode
	Signals []*IRNode
}
