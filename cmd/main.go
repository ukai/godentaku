// Copyright 2010 Fumitoshi Ukai. ALl Rights reserved.
// Use of this source code is governed by a BSD-style

// これはGoogle Developer Day 2010の「プログラミング言語Go」のセッションの
// 説明用に作ったプログラムです。
// 変数付き四則演算を行なうことができます。

// プログラムはmainパッケージのmain関数から実行されます。
package main

// 必要なパッケージをインポートします。
import (
	"bufio"
	"fmt"
	"os"
	// 標準パッケージではないので goinstall でインストールします。
	// goinstall godentaku.googlecode.com/hg/godentaku
	"godentaku.googlecode.com/hg/godentaku"
)

// inから一行づつ読んで、評価して表示します。
func readEvalPrint(in *bufio.Reader, env *godentaku.Env) (err os.Error) {
	// deferはこの関数をぬける時に実行するように指示します。
	// deferはこのようにpanic()/recover()につかったり、
	// リソースの開放などに使われます。
	// f, err := os.Open("file", os.O_RDONLY, 0666)
	// defer f.Close()
	// func() { .. } は関数リテラル/無名関数です。
	// func() { .. }() とすることでその関数の実行になります。
	// 引数があるばあいは func(i int){ .. }(10) のようにかきます。
	defer func() {
		// もし途中でpanic()していたらrecover()でつかまえることが
		// できます。
		if x := recover(); x != nil {
			fmt.Println("panic:", x)
		}
	}()

	// プロンプトを表示
	fmt.Printf(">")

	// '\n' まで(つまり1行)よみとり
	line, err := in.ReadBytes('\n')
	if err != nil {
		return err
	}
	ast, nbuf := godentaku.Read(line)
	v := godentaku.Eval(ast, env)
	fmt.Println(godentaku.Print(v, env))

	// Readで読みとらなかった残りが改行文字以外だったらwarning。
	if len(nbuf) > 0 && nbuf[0] != '\n' {
		fmt.Println("warning: unparsed:", string(nbuf))
	}
	return nil
}

// main関数がプログラムのエントリーです。
func main() {
	// 標準入力(os.Stdin)からbufioのReaderを作ります。
	in := bufio.NewReader(os.Stdin)

	// godentakuのEnvを作ります。
	env := godentaku.NewEnv()
	// godentakuのSetFunc関数を呼びだします。
	godentaku.SetFunc(env, "dump", godentaku.DumpAst)
	godentaku.SetFunc(env, "print", godentaku.PrintAst)

	// REPL: Read-eval-print loop
	// このようにforループをかくと無限ループになります。
	for {
		err := readEvalPrint(in, env)
		if err != nil {
			break
		}
	}
}
