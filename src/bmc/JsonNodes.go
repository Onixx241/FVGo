package bmc

type SlangAST struct {
	Design      Design       `json:"design"`
	Definitions []Definition `json:"definitions"`
}

type Design struct {
	Name    string       `json:"name"`
	Kind    string       `json:"kind"`
	Addr    int64        `json:"addr"`
	Members []MemberNode `json:"members"`
}

type MemberNode struct {
	Name           string           `json:"name"`
	Kind           string           `json:"kind"`
	Addr           int64            `json:"addr"`
	ProcedureKind  string           `json:"procedureKind,omitempty"`
	Type           string           `json:"type,omitempty"`
	LifeTime       string           `json:"lifetime,omitempty"`
	Direction      string           `json:"direction,omitempty"`
	Value          string           `json:"value,omitempty"`
	InternalSymbol string           `json:"internalSymbol,omitempty"`
	NetType        *NetTypeNode     `json:"netType,omitempty"`
	Body           *BodyNode        `json:"body,omitempty"`
	Initializer    *InitializerNode `json:"initializer,omitempty"`
	Assignment     *ExpressionNode  `json:"assignment,omitempty"`
}

type ItemsNode struct {
	Expressions []*ExpressionNode `json:"expressions,omitempty"`
	Statement   *StatementNode    `json:"stmt,omitempty"`
}

type BodyNode struct {
	Name       string            `json:"name"`
	Kind       string            `json:"kind"`
	Addr       int64             `json:"addr"`
	Members    []MemberNode      `json:"members"`
	Timing     *TimingNode       `json:"timing,omitempty"`
	Statement  *StatementNode    `json:"stmt,omitempty"`
	Conditions []ConditionalNode `json:"conditions,omitempty"`
	List       []*ListNode       `json:"list,omitempty"`
	Items      []*ItemsNode      `json:"items,omitempty"`
	IfTrue     *TruthNode        `json:"ifTrue,omitempty"`
	IfFalse    *TruthNode        `json:"ifFalse,omitempty"`
	Expression *ExpressionNode   `json:"expr,omitempty"`
}

type ListNode struct {
	Kind       string             `json:"kind"`
	Expression *ExpressionNode    `json:"expr,omitempty"`
	Conditions *[]ConditionalNode `json:"conditions,omitempty"`
	IfTrue     *TruthNode         `json:"ifTrue,omitempty"`
	IfFalse    *TruthNode         `json:"ifFalse,omitempty"`
	Assignment *ExpressionNode    `json:"assignment,omitempty"`
}

type NetTypeNode struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Addr int64  `json:"addr"`
	Type string `json:"type"`
}

type Definition struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Addr int64  `json:"addr"`
}
