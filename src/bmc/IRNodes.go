package bmc

type ComponentType int

const (
	InputType ComponentType = iota //clean this later
	OutputType
	StateType
	SignalType
	GateType
	ConstantType
	ValueType
	LocalParamType
	SwitchType
	OperatorType
)

type LogicalType int

const (
	NotGate LogicalType = iota
	AndGate
	OrGate
	ShiftLeftGate
	NilGate
)

type TimingNode struct {
	Kind       string         `json:"kind"`
	Expression ExpressionNode `json:"expr"`
	Edge       string         `json:"edge"`
}

type StatementNode struct {
	Kind       string             `json:"kind"`
	BlockKind  string             `json:"blockKind"`
	Conditions []*ConditionalNode `json:"conditions"`
	Body       *BodyNode          `json:"body"`
	Expression *ExpressionNode    `json:"expr"`
}

type ConditionalNode struct {
	Kind    string          `json:"kind"`
	Type    string          `json:"type"`
	Op      string          `json:"op"`
	Operand OperandNode     `json:"operand"`
	Expr    *ExpressionNode `json:"expr"`
}

type OperandNode struct {
	Kind        string          `json:"kind"`
	Type        string          `json:"type"`
	Symbol      *string         `json:"symbol"`
	Operand     *OperandNode    `json:"operand"`
	OperandExpr *ExpressionNode `json:"operandExpr,omitempty"`
	Value       *string         `json:"value"`
	Constant    *string         `json:"constant"`
	Op          string          `json:"op,omitempty"`
	Left        *ExpressionNode `json:"left,omitempty"`
	Right       *ExpressionNode `json:"right,omitempty"`
}

type TruthNode struct {
	Kind string         `json:"kind"`
	Expr ExpressionNode `json:"expr"`
	Body *BodyNode
}

type ExpressionNode struct { //acts like a general node sometimes maybe make a new general one?
	Kind            string             `json:"kind"`
	Type            string             `json:"type"`
	Symbol          string             `json:"symbol,omitempty"`
	Value           *string            `json:"value,omitempty"`
	Constant        *string            `json:"constant,omitempty"`
	Operand         *OperandNode       `json:"operand,omitempty"`
	OperandExpr     *ExpressionNode    `json:"operandExpr,omitempty"`
	Op              string             `json:"op,omitempty"`
	Left            *LeftNode          `json:"left,omitempty"`
	Right           *RightNode         `json:"right,omitempty"`
	Conditions      *[]ConditionalNode `json:"conditions,omitempty"`
	NonBlockingBool *bool              `json:"isNonBlocking,omitempty"`
}

type LeftNode struct {
	Kind       string             `json:"kind"`
	Type       string             `json:"type"`
	Symbol     string             `json:"symbol,omitempty"`
	Operand    *OperandNode       `json:"operand,omitempty"`
	Left       *LeftNode          `json:"left,omitempty"`
	Right      *RightNode         `json:"right,omitempty"`
	Op         string             `json:"op,omitempty"`
	Conditions *[]ConditionalNode `json:"conditions,omitempty"`
	Value      *string            `json:"value,omitempty"`
	Constant   *string            `json:"constant,omitempty"`
}

type RightNode struct {
	Kind       string             `json:"kind"`
	Type       string             `json:"type"`
	Symbol     string             `json:"symbol,omitempty"`
	Operand    OperandNode        `json:"operand,omitempty"`
	Constant   *string            `json:"constant,omitempty"`
	Left       *LeftNode          `json:"left,omitempty"`
	Right      *RightNode         `json:"right,omitempty"`
	Op         string             `json:"op,omitempty"`
	Conditions *[]ConditionalNode `json:"conditions,omitempty"`
	Value      *string            `json:"value,omitempty"`
}

type ExpressionStatement struct {
	Kind string         `json:"kind"`
	Expr ExpressionNode `json:"expr"`
}

type IRNode struct {
	Name      string
	Type      ComponentType
	Gate      *LogicalType
	Width     int
	Op        string
	Inputs    []*IRNode
	NextState *IRNode
}

type ConditionalResultNode struct {
	Inputs               []*ExpressionNode
	Kind                 string
	Symbol               string
	ResultantSignalCombs []IRNode
}

type ConstNode struct {
	Name        string
	Kind        string
	Width       int
	Value       string
	Type        ComponentType
	Initializer InitializerNode
}

type InitializerNode struct {
	Kind     string
	Type     string
	Constant string
	Operand  OperandNode
}

type ExprIR struct {
	Kind   string
	Op     string
	Type   string
	Width  int
	Symbol string
	Value  string
	Left   *ExprIR
	Right  *ExprIR
	Args   []*ExprIR
}

type GuardedAssign struct {
	Target      *IRNode
	Value       *ExprIR
	Guard       *ExprIR
	NonBlocking bool
	ProcessKind string
	Order       int
}

type ProcessIR struct {
	Kind        string
	Timing      *TimingNode
	Assignments []*GuardedAssign
}

type AssertionIR struct {
	Kind        string
	ClockEdge   string
	ClockSignal string
	Implication string
	Antecedent  *ExprIR
	DelayMin    int
	DelayMax    int
	Consequent  *ExprIR
}

type DesignGraph struct {
	Inputs             []*IRNode
	Outputs            []*IRNode
	States             []*IRNode
	Signals            []*IRNode
	Processes          []*TimingNode
	LocalParams        []*ConstNode
	ConditionalResults []*ConditionalResultNode
	Expressions        []*ExpressionNode
	ProcessIRs         []*ProcessIR
	AssertionIRs       []*AssertionIR
}
