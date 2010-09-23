package main

import (
	"bufio"
	"fmt"
	"os"
	"godentaku.googlecode.com/hg/godentaku"
)

func readEvalPrint(in *bufio.Reader, env *godentaku.Env) (err os.Error) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("panic:", x)
		}
	}()
	line, err := in.ReadBytes('\n')
	if err != nil {
		return err
	}
	ast, nbuf := godentaku.Read(line)
	v := godentaku.Eval(ast, env)
	fmt.Println(godentaku.Print(v, env))
	if len(nbuf) > 0 && nbuf[0] != '\n' {
		fmt.Println("warning: unparsed:", string(nbuf))
	}
	return nil
}

func main() {
	in := bufio.NewReader(os.Stdin)
	env := godentaku.NewEnv()
	godentaku.SetFunc(env, "dump", godentaku.DumpAst)
	godentaku.SetFunc(env, "print", godentaku.PrintAst)
	for {
		fmt.Printf(">")
		err := readEvalPrint(in, env)
		if err != nil {
			break
		}
	}
}
