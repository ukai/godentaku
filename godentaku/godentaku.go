package godentaku

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Ast interface {
	String() string
	Eval(*Env) Ast
}

type Env struct {
	Var map[string]Ast
	Func map[string]func(Ast, *Env) Ast
}

func NewEnv() *Env {
	env := &Env{}
	env.Var = make(map[string]Ast)
	env.Func = make(map[string]func(Ast, *Env) Ast)
	Set(env, ".printBase", 10)
	return env
}

func Set(env *Env, key string, n int) {
	env.Var[key] = Num{Val:n}
}

func SetExpr(env *Env, key string, expr Ast) {
	env.Var[key] = expr
}

func SetFunc(env *Env, funcname string, funcCode func(Ast, *Env) Ast) {
	env.Func[funcname] = funcCode
}

func envValue(env *Env, key string) (n int, ok bool) {
	if v, found := env.Var[key]; found {
		if n, ok := v.(Num); ok {
			return n.Val, true
		}
	}
	return 0, false
}

func Defined(env *Env, key string) bool {
	if n, found := envValue(env, key); found {
		return n != 0
	}
	return false
}

type Num struct {
	Val int
}
func (n Num) String() string {
	return strconv.Itoa(n.Val)
}
func (n Num) Eval(_ *Env) Ast {
	return n
}

type Symbol struct {
	Name string
}
func (s Symbol) String() string {
	return s.Name
}
func (s Symbol) Eval(env *Env) Ast {
	if v, found := env.Var[s.Name]; found {
		env.Var[s.Name] = Num{Val:0}  // prevent recursive eval
		r := v.Eval(env)
		env.Var[s.Name] = v
		return r
	}
	return s
}

type UnaryOp struct {
	Op byte
	Expr Ast
}
func (e UnaryOp) String() string {
	return fmt.Sprintf("%c%s", e.Op, e.Expr)
}
func (e UnaryOp) Eval(env *Env) Ast {
	if e.Op != 0 && e.Op != '-' {
		panic(fmt.Sprintf("unsupported uniOp:%c", e.Op))
	}
	v := e.Expr.Eval(env)
	if n, ok := v.(Num); ok && e.Op == '-' {
		return Num{Val: -n.Val}
	}
	return v
}

type BinOp struct {
	Op byte
	Left Ast
	Right Ast
}
func (e BinOp) String() string {
	return fmt.Sprintf("(%s %c %s)", e.Left, e.Op, e.Right)
}
func (e BinOp) Eval(env *Env) Ast {
	l := e.Left.Eval(env)
	r := e.Right.Eval(env)

	lnum, lok := l.(Num)
	rnum, rok := r.(Num)
	if lok && rok {
		switch e.Op {
		case '+': return Num{Val:lnum.Val + rnum.Val}
		case '-': return Num{Val:lnum.Val - rnum.Val}
		case '*': return Num{Val:lnum.Val * rnum.Val}
		case '/': return Num{Val:lnum.Val / rnum.Val}
		}
		panic(fmt.Sprintf("unsupported binOp:%c", e.Op))
	}
	return BinOp{Op:e.Op, Left:l, Right:r}
}

type AssignOp struct {
	Var Symbol
	Expr Ast
}
func (a AssignOp) String() string {
	return fmt.Sprintf("%s = %s", a.Var, a.Expr)
}
func (a AssignOp) Eval(env *Env) Ast {
	v := a.Expr.Eval(env)
	if s, ok := a.Expr.(Symbol); ok && s.Name == "undef" {
		env.Var[a.Var.Name] = a.Expr, false
	} else {
		env.Var[a.Var.Name] = a.Expr
	}
	return v
}

type FunCall struct {
	Func Symbol
	Expr Ast
}
func (f FunCall) String() string {
	return fmt.Sprintf("%s(%s)", f.Func, f.Expr)
}
func (f FunCall) Eval(env *Env) Ast {
	if fun, ok := env.Func[f.Func.Name]; ok {
		return fun(f.Expr, env)
	}
	panic(fmt.Sprintf("no such function: %s", f.Func.Name))
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func digitVal(b byte) int {
	switch {
	case isDigit(b):
		return int(b - '0')
	case 'a' <= b && b <= 'f':
		return int(b - 'a' + 10)
	case 'A' <= b && b <= 'F':
		return int(b - 'A' + 10)
	}
	return -1
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}
func isSpace(b byte) bool {
	return (b == ' ' || b == '\t')
}
func skipSpace(buf []byte) []byte {
	for i := 0; i < len(buf); i++ {
		if !isSpace(buf[i]) {
			return buf[i:]
		}
	}
	return buf[len(buf):]
}

func getNum(buf []byte) (num Num, nbuf []byte) {
	if !isDigit(buf[0]) {
		panic("not number:" + string(buf))
	}
	n := int(buf[0] - '0')
	nbuf = buf[1:]
	base := 10
	if n == 0 {
		switch nbuf[0] {
		case 'b', 'B':base = 2; nbuf = nbuf[1:]
		case 'x', 'X': base = 16; nbuf = nbuf[1:]
		default:
			if isDigit(nbuf[0]) {
				base = 8
			}
		}
	}
	for len(nbuf) > 0 {
		if d := digitVal(nbuf[0]); d >= 0 && d < base {
			n = n * base + d
		} else {
			break
		}
		nbuf = nbuf[1:]
	}
	return Num{Val: n}, nbuf
}

func getSymbol(buf []byte) (sym Symbol, nbuf []byte) {
	if !isAlpha(buf[0]) && buf[0] != '.' {
		panic("not symbol:" + string(buf))
	}
	var i int
	for i = 1; i < len(buf); i++ {
		if !isAlpha(buf[i]) && !isDigit(buf[i]) {
			break
		}
	}
	return Symbol{Name: string(buf[0:i])}, buf[i:]
}

// stmt := expr | symbol '=' expr
// expr := [+|-] term ([+|-] term)
// term := factor ([*|/] factor)
// factor := num | symbol | '(' expr ')' | symbol'(' expr ')'

func parseStatement(buf []byte) (stmt Ast, nbuf []byte) {
	stmt, nbuf = parseExpression(buf)
	nbuf = skipSpace(nbuf)
	if nbuf[0] == '=' {
		if sym, ok := stmt.(Symbol); ok {
			var expr Ast
			expr, nbuf = parseExpression(nbuf[1:])
			stmt = AssignOp{Var:sym, Expr:expr}
		} else {
			panic(fmt.Sprintf("lvalue is not symbol:%s", stmt))
		}
	}
	return stmt, nbuf
}

func parseExpression(buf []byte) (expr Ast, nbuf []byte) {
	buf = skipSpace(buf)
	var uniop byte
	if buf[0] == '+' || buf[0] == '-' {
		uniop = buf[0]
		buf = buf[1:]
	}
	expr, nbuf = parseTerm(buf)
	if uniop == '-' {
		expr = UnaryOp{Op:'-', Expr:expr}
	}
	nbuf = skipSpace(nbuf)
	for nbuf[0] == '+' || nbuf[0] == '-' {
		op := nbuf[0]
		var term Ast
		term, nbuf = parseTerm(nbuf[1:])
		expr = BinOp{Op: op, Left: expr, Right: term}
		nbuf = skipSpace(nbuf)
	}
	return expr, nbuf
}
func parseTerm(buf []byte) (term Ast, nbuf []byte) {
	term, nbuf = parseFactor(buf)
	nbuf = skipSpace(nbuf)
	for nbuf[0] == '*' || nbuf[0] == '/' {
		op := nbuf[0]
		var factor Ast
		factor, nbuf = parseFactor(nbuf[1:])
		term = BinOp{Op: op, Left: term, Right: factor}
		nbuf = skipSpace(nbuf)
	}
	return term, nbuf
}

func parseFactor(buf []byte) (factor Ast, nbuf []byte) {
	buf = skipSpace(buf)
	switch ch := buf[0]; {
	case ch == '(':
		factor, nbuf = parseExpression(buf[1:])
		nbuf = skipSpace(nbuf)
		if nbuf[0] != ')' {
			panic("unbalanced paren: " + string(buf))
		}
		return factor, nbuf[1:]
	case isDigit(ch):
		return getNum(buf)
	case isAlpha(ch) || ch == '.':
		var sym Symbol
		sym, nbuf = getSymbol(buf)
		nbuf = skipSpace(nbuf)
		if nbuf[0] == '(' {
			var expr Ast
			expr, nbuf = parseExpression(nbuf[1:])
			nbuf = skipSpace(nbuf)
			if nbuf[0] != ')' {
				panic("unbalanced paren for func:" + sym.Name)
			}
			return FunCall{Func:sym, Expr:expr}, nbuf[1:]
		}
		return sym, nbuf
	}
	panic("unepxected token:" + string(buf))
}

func Read(in *bufio.Reader) (ast Ast, err os.Error) {
	line, err := in.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	ast, _ = parseStatement(line)
	return ast, nil
}

func Eval(ast Ast, env *Env) (v Ast) {
	v = ast.Eval(env)
	if s, ok := ast.(Symbol); ok && s.Name == "_" {
	} else {
		SetExpr(env, "_", ast)
	}
	return v
}

func Print(v Ast, env *Env) string {
	if n, ok := v.(Num); ok {
		var format string
		base, _ := envValue(env, ".printBase")
		switch base {
		case 2: format = "0b%b"
		case 8: format = "0%o"
		case 10: format = "%d"
		case 16: format = "0x%x"
		default:
			panic(fmt.Sprintf("bad .printBase: %d",
				env.Var[".printBase"]))
		}
		return fmt.Sprintf(format, n.Val)
	}
	return v.String()
}

func DumpAst(v Ast, env *Env) Ast {
	if s, ok := v.(Symbol); ok {
		if e, found := env.Var[s.Name]; found {
			return Symbol{Name: fmt.Sprintf("%#v", e)}
		}
	}
	return Symbol{Name: fmt.Sprintf("%#v", v)}
}

func PrintAst(v Ast, env *Env) Ast {
	if s, ok := v.(Symbol); ok {
		if e, found := env.Var[s.Name]; found {
			return Symbol{Name: e.String()}
		}
	}
	return Symbol{Name: v.String()}
}
