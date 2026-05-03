package parser

type Node interface{ astNode() }

type Expression interface {
	Node
	exprNode()
}

type NumberLiteral struct{ Value float64 }
type StringLiteral struct{ Value string }
type BoolLiteral struct{ Value bool }
type NullLiteral struct{}
type Identifier struct{ Name string }
type ListLiteral struct{ Elements []Expression }

type BinaryExpr struct {
	Op    string
	Left  Expression
	Right Expression
}

type UnaryExpr struct {
	Op      string
	Operand Expression
}

type CallExpr struct {
	Callee Expression
	Args   []Expression
}

type IndexExpr struct {
	Collection Expression
	Index      Expression
}

func (NumberLiteral) astNode() {}
func (StringLiteral) astNode() {}
func (BoolLiteral) astNode()   {}
func (NullLiteral) astNode()   {}
func (Identifier) astNode()    {}
func (ListLiteral) astNode()   {}
func (BinaryExpr) astNode()    {}
func (UnaryExpr) astNode()     {}
func (CallExpr) astNode()      {}
func (IndexExpr) astNode()     {}

func (NumberLiteral) exprNode() {}
func (StringLiteral) exprNode() {}
func (BoolLiteral) exprNode()   {}
func (NullLiteral) exprNode()   {}
func (Identifier) exprNode()    {}
func (ListLiteral) exprNode()   {}
func (BinaryExpr) exprNode()    {}
func (UnaryExpr) exprNode()     {}
func (CallExpr) exprNode()      {}
func (IndexExpr) exprNode()     {}

// ----- Statements -----

type Statement interface {
	Node
	stmtNode()
}

type VarDecl struct {
	Name    string
	Value   Expression
	IsConst bool
}

type AssignStmt struct {
	Name  string
	Op    string // "=", "+=", "-=", "*=", "/="
	Value Expression
}

type IfStmt struct {
	Cond Expression
	Then *BlockStmt
	Else Statement
}

type WhileStmt struct {
	Cond Expression
	Body *BlockStmt
}

type ForStmt struct {
	VarName   string
	Iterable  Expression
	RangeArgs []Expression
	Body      *BlockStmt
}

type FuncDecl struct {
	Name   string
	Params []string
	Body   *BlockStmt
}

type ReturnStmt struct {
	Value Expression
}

type BreakStmt struct{}
type ContinueStmt struct{}

type ExprStmt struct {
	Expr Expression
}

type BlockStmt struct {
	Statements []Statement
}

func (VarDecl) astNode()      {}
func (AssignStmt) astNode()   {}
func (IfStmt) astNode()       {}
func (WhileStmt) astNode()    {}
func (ForStmt) astNode()      {}
func (FuncDecl) astNode()     {}
func (ReturnStmt) astNode()   {}
func (BreakStmt) astNode()    {}
func (ContinueStmt) astNode() {}
func (ExprStmt) astNode()     {}
func (BlockStmt) astNode()    {}

func (VarDecl) stmtNode()      {}
func (AssignStmt) stmtNode()   {}
func (IfStmt) stmtNode()       {}
func (WhileStmt) stmtNode()    {}
func (ForStmt) stmtNode()      {}
func (FuncDecl) stmtNode()     {}
func (ReturnStmt) stmtNode()   {}
func (BreakStmt) stmtNode()    {}
func (ContinueStmt) stmtNode() {}
func (ExprStmt) stmtNode()     {}
func (BlockStmt) stmtNode()    {}

type Program struct {
	Statements []Statement
}

func (Program) astNode() {}
