package test

import (
	"fmt"
	"testing"

	"github.com/RainyBow/gotcl"
)

func TestCase1(t *testing.T) {
	ir, err := gotcl.NewInterpreter(`
	lappend auto_path /opt/Spirent_TestCenter_5.29/Spirent_TestCenter_Application_Linux
	`)
	if err == nil {
		if _, err := ir.EvalAsString("puts $auto_path"); err != nil {
			t.Fail()
		}
		fmt.Println(ir.EvalAsString("package require stc"))
	} else {
		t.Fail()
	}

}
func TestCase2(t *testing.T) {
	_, err := gotcl.NewInterpreter(`
	puts $abcd
	`)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Fail()

}
