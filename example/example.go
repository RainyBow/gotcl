package main

import (
	"fmt"

	"github.com/RainyBow/gotcl"
)

func main() {
	ir, err := gotcl.NewInterpreter("")
	if err != nil {
		panic(err)
	}
	a, err := ir.EvalAsInt("set a [expr 1 + 2]")
	// a = 3 , err = <nil>
	fmt.Printf("a = %d ,err = %v\n", a, err)

}
