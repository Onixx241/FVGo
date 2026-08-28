package slang

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
	Name           string      `json:"name"`
	Kind           string      `json:"kind"`
	Addr           int64       `json:"addr"`
	Type           string      `json:"type,omitempty"`
	Direction      string      `json:"direction,omitempty"`
	InternalSymbol string      `json:"internalSymbol,omitempty"`
	NetType        NetTypeNode `json:"netType,omitempty"`
	Body           BodyNode    `json:"body,omitempty"`
}

type BodyNode struct {
	Name    string       `json:"name"`
	Kind    string       `json:"kind"`
	Addr    int64        `json:"addr"`
	Members []MemberNode `json:"members"`
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
