// Copyright 2010 Fumitoshi Ukai. ALl Rights reserved.
// Use of this source code is governed by a BSD-style

// これはGoogle Developer Day 2010の「プログラミング言語Go」のセッションの
// 説明用に作ったプログラムです。
// 変数付き四則演算を行なうことができます。
// Go言語の基本的な機能を説明することを目的としていたので、便利なpackageは
// できるだけ使わないで書くことにしてあります。

// package文で、パッケージ名を指定します。このパッケージ名がimportした時の
// defaultのパッケージ名となります。
//  import "godentaku.googlecode.com/hg/godentaku"
// 違うパッケージ名で参照したい場合は
//   import dentaku "godentaku.googlecode.com/hg/godentaku"
// のようにします。
// 通常はパッケージ名と同じディレクトリ名の中にそのパッケージに所属する
// ソースファイルをまとめていれておいてビルドします。
package godentaku

// このソースで利用しているパッケージをインポートします。インポートしない
// と利用することができません。
// 間接的に使われるパッケージに関してはインポートする必要はありません。
// 利用していないパッケージをインポートするのはコンパイルエラーです。
// インポートによる副作用が必要な場合は
//  import _ "some/package"
// のように _ としてインポートします。
import "fmt"

// 型定義です。
// string型を返すString()というメソッドとEnv型へのポインタをうけとってAst型を
// 返すEval()というメソッドをもつインターフェイスを Astという型として定義
// しています。
// Ast型の変数は、(nilでなければ)これらのメソッドをかならずもっていて
// 呼びだすことができます。
// Ast型の変数に代入できるオブジェクトは、これらのメソッドをもった型である
// 必要があります。
// 大文字ではじまっているので、この型はパッケージの外で利用できます。
type Ast interface {
	String() string
	Eval(*Env) Ast
}

// 型定義です。
// string型をキーにしてAst型の要素をもつmap型のVarというフィールドと
// string型をキーにして、Ast型とEnv型へのポインタをうけとってAst型を返す関数を
// 要素としてもつmap型のFuncというフィールドをもつ構造体を Env型として定義
// しています。
// mapのキーにはstruct, array, slice型は使えません。
// 大文字ではじまっているので、この型およびこの型の中のフィールドは
// パッケージの外で利用できます。
type Env struct {
	Var  map[string]Ast
	Func map[string]func(Ast, *Env) Ast
}

// NewEnv()という関数定義です。
// Env型へのポインタをかえします。
// 大文字ではじまっているので、この関数はパッケージの外で呼び出すことが
// できます。
func NewEnv() *Env {
	// Env型の領域を確保して envにそのポインタをセットします。
	// env := new(Env) と書くこともできます。
	// := は式の型をもった変数を宣言・初期化することなので
	// var env Env = new(Env) とおなじです。
	// 要素の内容は0初期化されています。
	env := &Env{}
	// mapはビルトイン参照型なのでmakeで初期化する必要があります。
	// makeで初期化するのはmapの他にslice, chanがあります。
	env.Var = make(map[string]Ast)
	env.Func = make(map[string]func(Ast, *Env) Ast)
	// 以上の3行は次のように書くこともできます。
	//  env := &Env{Var: make(map[string]Ast),
	//              Func: make(map[string]func(Ast,*Env)Ast)}

	// Set関数の呼出です。定義が後にあっても大丈夫です。
	Set(env, ".printBase", 10)
	return env
}

// Set()という関数定義です。
// Env型へのポインタと、string型とint型をうけとって処理します。
// 返り値はありません。
func Set(env *Env, key string, n int) {
	// mapにキーに対する要素の設定は次のように書けます。
	env.Var[key] = Num(n)
	// ちなみに要素を消すときは次のように書きます。
	// env.Var[key] = Num(0), false
}

// SetExpr()という関数定義です。
// Goではfunction overloadはできません。
func SetExpr(env *Env, key string, expr Ast) {
	env.Var[key] = expr
}

// SetFunc()という関数定義です。
// 関数(func)も普通の値と同じようにあつかうことができます。
func SetFunc(env *Env, funcname string, funcCode func(Ast, *Env) Ast) {
	env.Func[funcname] = funcCode
}

// envValue()という関数定義です。
// Env型へのポインタとstring型をうけとって、int型とbool型をかえします。
// このように多値をかえすことは普通におこなえます。
// 返り値のパラメータも変数として使うことができます。
// 小文字ではじまっているので、この関数はパッケージの外では見えません。
func envValue(env *Env, key string) (n int, ok bool) {
	// mapの中にキーにたいする値があるかどうか調べるにはこのように
	// 書きます。found がtrueなら要素あり、falseならなしです。
	// v := env.Var[key]としてkeyに対する値がない時は要素型の初期値
	// がかえってきます。
	if v, found := env.Var[key]; found {
		// v は Varの型定義によりAst型の変数です。
		// num, ok := v.(Num)とよびだすことで、Num型へ型変換をためす
		// ことができます。型変換できればnumはNum型になったときの値、
		// okがtrueになりませす。型変換できなければnumはNum型の初期値、
		// okがfalseです。
		if num, ok := v.(Num); ok {
			// numは下記のようにもともとint型なのでint型に変換
			// できます。
			return int(num), true
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

// 型定義です。
// int型とおなじデータをNum型にできます。このようにして定義した型にはメソッドを
// 定義できるようになります。
// int型をNum型にするには Num(i)、Num型をint型にするにはint(num)のようにします。
// Num型の変数はメモリ上ではint型の変数とおなじです。
type Num int

// メソッド定義です。
// Num型のオブジェクトにstring型を返すString()というメソッドを定義しています。
// いわゆるC++のthisにあたるのがメソッド名の前にきます。ここで定義した
// 変数がthis扱いになります。
func (n Num) String() string {
	// Num型のnをint型に変換して、それをfmt.Sprintf()の"%d"で文字列に
	// 変換してかえしています。
	// strconvパッケージを使ってstrconv.Itoa(int(n))とも書けます。
	// string(int(n))はそのint値をunicode codepointとなるUTF-8文字列
	// になってしまいます。
	return fmt.Sprintf("%d", int(n))
}
// メソッド定義です。
// Num型のオブジェクトにEnv型へのポインタをうけとって、Ast型をかえすEval()という
// メソッドを定義しています。
// 引数のEnv型へのポインタは利用しないので _ という変数を使います。
func (n Num) Eval(_ *Env) Ast {
	// Num型は、Astというインターフェイス型に必要なString(), Eval()
	// メソッドをもっているのでそのままAst型に代入・変換できます。
	return n
}
// 以上でNum型はAstインターフェイスをみたしていることになります。
// 明示的にコンパイルチェックしたい場合は次のように書いてもかまいません。
// var _ Ast = Num(0)
// もしNum型がAstインターフェイスをみたしていなければ、この代入はコンパイル
// エラーになります。代入先が _ なので代入するという操作自体は実行されません。

// string型をAstインターフェイスをみたすSymbol型として定義します。
type Symbol string

func (s Symbol) String() string {
	// Symbol型のsをstring型にするにはstring(s)とします。
	return string(s)
}
func (s Symbol) Eval(env *Env) Ast {
	// Symbolを評価した結果をかえします。
	name := string(s)
	// もしSymbol名がVarに登録されていたらその値に展開します。
	if v, found := env.Var[name]; found {
		// 再帰的に評価されるのをふせぐために0にしています。
		// もしこれがないと a = a + 1 のような式がどうなるか
		// 考えてみましょう
		env.Var[name] = Num(0)
		// Symbolに代入されていた式を評価します。
		r := v.Eval(env)
		// 元に戻します。
		env.Var[name] = v
		return r
	}
	// もしVarになければそのままかえします。
	return s
}

// 単項式をAstインターフェイスをみたすUnaryOp型として定義します。
type UnaryOp struct {
	Op   byte
	Expr Ast
}

func (e UnaryOp) String() string {
	return fmt.Sprintf("%c%s", e.Op, e.Expr)
}
func (e UnaryOp) Eval(env *Env) Ast {
	if e.Op != '-' {
		// もし単項演算子が '-' じゃなかったpanicします。
		// 処理をうちきって呼出元にもどっていきます。
		// 途中 recoverされれば、ここで渡した値がとりだせます。
		// recoverされなければ、プログラムは異常終了します。
		panic(fmt.Sprintf("unsupported uniOp:%c", e.Op))
	}
	// UnaryOpのExprフィールドの内容を Evalします。
	v := e.Expr.Eval(env)
	// もし評価した結果が Num型だったら、マイナスにした値にしてかえします。
	if n, ok := v.(Num); ok && e.Op == '-' {
		return Num(-int(n))
	}
	return v
}

// 二項式をAstインターフェイスをみたすBinOp型として定義します。
type BinOp struct {
	Op    byte
	Left  Ast
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
	// 左辺値、右辺値 評価して両方Num型だったら計算した結果にして
	// かえします。
	if lok && rok {
		switch e.Op {
		case '+':
			return Num(int(lnum) + int(rnum))
		case '-':
			return Num(int(lnum) - int(rnum))
		case '*':
			return Num(int(lnum) * int(rnum))
		case '/':
			return Num(int(lnum) / int(rnum))
		}
		panic(fmt.Sprintf("unsupported binOp:%c", e.Op))
	}
	// 左辺値、右辺値を評価した結果にしたBinOpをつくってかえします。
	return BinOp{Op: e.Op, Left: l, Right: r}
}

// 代入式をAstインターフェイスをみたすBinOp型として定義します。
type AssignOp struct {
	Var  Symbol
	Expr Ast
}

func (a AssignOp) String() string {
	return fmt.Sprintf("%s = %s", a.Var, a.Expr)
}
func (a AssignOp) Eval(env *Env) Ast {
	v := a.Expr.Eval(env)
	// もし"undef"という式を代入する場合は、VarからSymbolの情報を削除します
	if s, ok := a.Expr.(Symbol); ok && string(s) == "undef" {
		// , falseをわたすことでmapから消すことができます。
		env.Var[string(a.Var)] = a.Expr, false
	} else {
		env.Var[string(a.Var)] = a.Expr
	}
	return v
}

// 関数呼出をAstインターフェイスをみたすBinOp型として定義します。
type FunCall struct {
	Func Symbol
	Expr Ast
}

func (f FunCall) String() string {
	return fmt.Sprintf("%s(%s)", f.Func, f.Expr)
}
func (f FunCall) Eval(env *Env) Ast {
	// f.FuncというSymbolがenv.Funcにあったらその関数を呼出ます。
	// funはenv.Funcの定義によりfunc (Ast, *Env) Astです。
	if fun, ok := env.Func[string(f.Func)]; ok {
		return fun(f.Expr, env)
	}
	panic(fmt.Sprintf("no such function: %s", f.Func))
}

// bというbyte(ASCII文字)が数字かどうか 
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
// bというASCII文字を数値に変換します。
func digitVal(b byte) int {
	// switch は次のようにcaseに式を書くこともできます。
	// これは
	// if true == isDigit(b) { .. }
	// else if true == ('a' <= b && b <= 'f') { .. }
	// else if true == ('A' <= b && b <= 'F') { .. }
	// とおなじです。
	switch {
	case isDigit(b):
		return int(b - '0')
	case 'a' <= b && b <= 'f': // 16進用
		return int(b - 'a' + 10)
	case 'A' <= b && b <= 'F': // 16進用
		return int(b - 'A' + 10)
	}
	return -1
}
// bというASCII文字がアルファベットかどうか。_ もふくめています。
func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}
// bというASCII文字が空白文字かどうか
func isSpace(b byte) bool {
	return (b == ' ' || b == '\t')
}
// byte sliceをスキャンしていって空白文字をスキップします。
func skipSpace(buf []byte) []byte {
	// byte sliceはbyte配列のようにアクセスできるのでこのようなfor文で
	// スキャンすることができます。
	// sliceの長さもlen関数をつかいます。
	for i := 0; i < len(buf); i++ {
		if !isSpace(buf[i]) {
			// 空白文字じゃないのがみつかったら、その場所以降を
			// byte sliceとしてかえします。
			// : の後ろを省略した場合、残り全部になります。
			return buf[i:]
		}
	}
	// 空のsliceは次のように書けます。
	return []byte{}
}

// byte sliceをスキャンして数字をNum型としてとりだします。
// nbufは次にスキャンしていくところをさします。
// 単に文字列を数字にするなら fmt.SScanf(), strconv.Atoi()などがあります。
// scannerというパッケージもあります。
func getNum(buf []byte) (num Num, nbuf []byte) {
	if !isDigit(buf[0]) {
		// 先頭が数字じゃなければ panicします。
		// byte sliceを文字列にするには string(buf)とします。
		// stringの中身はsliceとちがって変更することができません。
		// s := "hello"; s[0] = 'H' はエラーです。
		// s += ", world" はできます。
		panic("not number:" + string(buf))
	}
	n := int(buf[0] - '0')
	nbuf = buf[1:] // 1バイトすすめます。
	base := 10
	if n == 0 {
		// switchはこのように書くこともできます。
		switch nbuf[0] {
		// 0b.., 0B... の場合
		case 'b', 'B':
			base = 2
			nbuf = nbuf[1:]
		// baseを2にしてnbufを1バイトすすめます。
		// Cなどとちがってcaseは次のcaseとはつながっていません。
		// (わざわざbreakをかかない)
		// どうしてもつなげたい場合は fallthrough とかきます。

		// 0x.., 0X... の場合
		case 'x', 'X':
			base = 16
			nbuf = nbuf[1:]
		default:
			if isDigit(nbuf[0]) {
				base = 8
			}
		}
	}
	// for文はwhileのような書きかたもできます。(whileはありません)
	for len(nbuf) > 0 {
		if d := digitVal(nbuf[0]); d >= 0 && d < base {
			n = n*base + d
		} else {
			break
		}
		nbuf = nbuf[1:]
	}
	return Num(n), nbuf
}

// byte sliceをスキャンして文字列をSymbol型としてとりだします。
func getSymbol(buf []byte) (sym Symbol, nbuf []byte) {
	if !isAlpha(buf[0]) && buf[0] != '.' {
		panic("not symbol:" + string(buf))
	}
	// for i := 1; i < len(buf); i++ { .. } とかくと i のスコープは
	// forの中だけになってしまいます。
	var i int
	for i = 1; i < len(buf); i++ {
		if !isAlpha(buf[i]) && !isDigit(buf[i]) {
			break
		}
	}
	// 0〜iを文字列にしてSymbol型に、i以降を次にスキャンすべきsliceに
	// します。
	return Symbol(string(buf[0:i])), buf[i:]
}

// 四則演算の簡単な再帰降下パーザです。
// stmt := expr '\n' | symbol '=' expr '\n'
// expr := [+|-] term ([+|-] term)
// term := factor ([*|/] factor)
// factor := num | symbol | '(' expr ')' | symbol'(' expr ')'
// より複雑な文法はgoyaccなどを使ったほうがいいでしょう。
// goパッケージがgoのパーザを含んでいるのでそれも参考になります。

// stmt := expr '\n' | symbol '=' expr '\n'
// をbufからよんで、stmtをあらわすAstと、次のよむsliceをかえします。
func parseStatement(buf []byte) (stmt Ast, nbuf []byte) {
	buf = skipSpace(buf)
	stmt, nbuf = parseExpression(buf)
	nbuf = skipSpace(nbuf)
	if nbuf[0] == '=' {
		// もし symbol '=' の場合
		if sym, ok := stmt.(Symbol); ok {
			var expr Ast
			expr, nbuf = parseExpression(nbuf[1:])
			// 代入式としてあつかいます。
			stmt = AssignOp{Var: sym, Expr: expr}
		} else {
			// '=' の左はSymbol以外だと例外処理にします。
			panic(fmt.Sprintf("lvalue is not symbol:%s", stmt))
		}
	}
	return stmt, nbuf
}

// expr := [+|-] term ([+|-] term)
// をbufからよんで、exprをあらわすAstと、次のよむsliceをかえします。
func parseExpression(buf []byte) (expr Ast, nbuf []byte) {
	buf = skipSpace(buf)
	var uniop byte
	if buf[0] == '+' || buf[0] == '-' {
		uniop = buf[0] // '+' か '-'
		buf = buf[1:]  // 1バイトすすめる
	}
	expr, nbuf = parseTerm(buf)
	if uniop == '-' {
		// '-' term だったら UnaryOpをつくる
		// '+' term は termとおなじなのでなにもしません。
		expr = UnaryOp{Op: '-', Expr: expr}
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

// term := factor ([*|/] factor)
// をbufからよんで、termをあらわすAstと、次のよむsliceをかえします。
func parseTerm(buf []byte) (term Ast, nbuf []byte) {
	buf = skipSpace(buf)
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

// factor := num | symbol | '(' expr ')' | symbol'(' expr ')'
// をbufからよんで、factorをあらわすAstと、次のよむsliceをかえします。
func parseFactor(buf []byte) (factor Ast, nbuf []byte) {
	buf = skipSpace(buf)
	// switch は次のように書くこともできます。
	switch ch := buf[0]; {
	case ch == '(': // '(' expr ')' の場合
		factor, nbuf = parseExpression(buf[1:])
		nbuf = skipSpace(nbuf)
		if nbuf[0] != ')' {
			panic("unbalanced paren: " + string(buf))
		}
		return factor, nbuf[1:]
	case isDigit(ch): // 数字の場合
		return getNum(buf)
	case isAlpha(ch) || ch == '.': // symbolの場合
		var sym Symbol
		sym, nbuf = getSymbol(buf)
		nbuf = skipSpace(nbuf)
		if nbuf[0] == '(' { // symbol '(' expr ')' の場合
			var expr Ast
			expr, nbuf = parseExpression(nbuf[1:])
			nbuf = skipSpace(nbuf)
			if nbuf[0] != ')' {
				panic("unbalanced paren for func:" + string(sym))
			}
			// FunCallを作ります。
			return FunCall{Func: sym, Expr: expr}, nbuf[1:]
		}
		return sym, nbuf
	}
	panic("unepxected token:" + string(buf))
}

// byte sliceを読んでAst型にします。
// 大文字ではじまっているのでパッケージの外から呼びだせます。
func Read(b []byte) (ast Ast, nbuf []byte) {
	return parseStatement(b)
}

// Astを評価してAstをかえします。
// 大文字ではじまっているのでパッケージの外から呼びだせます。
func Eval(ast Ast, env *Env) (v Ast) {
	v = ast.Eval(env)
	if s, ok := ast.(Symbol); ok && string(s) == "_" {
	} else {
		// さっきの式を _ で参照できるようにセットしておきます。
		SetExpr(env, "_", ast)
	}
	return v
}

// Astを文字列にします。
// 大文字ではじまっているのでパッケージの外から呼びだせます。
func Print(v Ast, env *Env) string {
	// Num型の場合 .printBaseの値によって基数をかえます。
	if n, ok := v.(Num); ok {
		var format string
		// もし多値をかえす関数で使わない返り値があるときは
		// _ でうけとります。
		base, _ := envValue(env, ".printBase")
		switch base {
		case 2:
			format = "0b%b"
		case 8:
			format = "0%o"
		case 10:
			format = "%d"
		case 16:
			format = "0x%x"
		default:
			panic(fmt.Sprintf("bad .printBase: %d",
				env.Var[".printBase"]))
		}
		return fmt.Sprintf(format, int(n))
	}
	return v.String()
}

func DumpAst(v Ast, env *Env) Ast {
	if s, ok := v.(Symbol); ok {
		if e, found := env.Var[string(s)]; found {
			return Symbol(fmt.Sprintf("%#v", e))
		}
	}
	// %#v を使うと型情報つきで pretty printできます。
	return Symbol(fmt.Sprintf("%#v", v))
}

func PrintAst(v Ast, env *Env) Ast {
	if s, ok := v.(Symbol); ok {
		if e, found := env.Var[string(s)]; found {
			return Symbol(e.String())
		}
	}
	return Symbol(v.String())
}
