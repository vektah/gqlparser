package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/tools/imports"
)

func main() {
	out := bytes.Buffer{}
	out.WriteString("package validator\n\n")

	file, err := ioutil.ReadFile("prelude.graphql")
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(&out, `var Prelude = &ast.Source{
		Name: "prelude.graphql",
		Input: %q,
		BuiltIn: true,
	}`, string(file))

	formatted, err2 := imports.Process("prelude.go", out.Bytes(), nil)
	if err2 != nil {
		panic(err2)
	}

	ioutil.WriteFile("prelude.go", formatted, 0644)
}
