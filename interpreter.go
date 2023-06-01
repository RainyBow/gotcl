package gotcl

/*
#cgo !tcl85 LDFLAGS: -ltcl8.6
#cgo !tcl85 CFLAGS: -I/usr/include/tcl8.6
#cgo tcl85 LDFLAGS: -ltcl8.5
#cgo tcl85 CFLAGS: -I/usr/include/tcl8.5
#include <stdlib.h>
#include <tcl.h>
*/
import "C"
import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type interpreter struct {
	C      *C.Tcl_Interp
	thread C.Tcl_ThreadId
	cmdbuf bytes.Buffer
	lock   *sync.Mutex
}

func new_interpreter() (*interpreter, error) {
	ir := &interpreter{
		C:      C.Tcl_CreateInterp(),
		thread: C.Tcl_GetCurrentThread(),
		lock:   &sync.Mutex{},
	}
	status := C.Tcl_Init(ir.C)
	if status != C.TCL_OK {
		return nil, errors.New(C.GoString(C.Tcl_GetStringResult(ir.C)))
	}
	return ir, nil
}

func (ir *interpreter) eval(script []byte) error {
	if len(script) == 0 {
		return nil
	}
	status := C.Tcl_EvalEx(ir.C, (*C.char)(unsafe.Pointer(&script[0])),
		C.int(len(script)), 0)
	if status != C.TCL_OK {
		return errors.New(C.GoString(C.Tcl_GetStringResult(ir.C)))
	}
	return nil
}

func (ir *interpreter) eval_as(out interface{}, script []byte) error {
	pv := reflect.ValueOf(out)
	if pv.Kind() != reflect.Ptr || pv.IsNil() {
		panic("gothic: EvalAs expected a non-nil pointer argument")
	}
	v := pv.Elem()

	err := ir.eval(script)
	if err != nil {
		return err
	}

	return ir.tcl_obj_to_go_value(C.Tcl_GetObjResult(ir.C), v)
}
func (ir *interpreter) tcl_obj_to_go_value(obj *C.Tcl_Obj, v reflect.Value) error {
	var status C.int

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var out C.Tcl_WideInt
		status = C.Tcl_GetWideIntFromObj(ir.C, obj, &out)
		if status == C.TCL_OK {
			v.SetInt(int64(out))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var out C.Tcl_WideInt
		status = C.Tcl_GetWideIntFromObj(ir.C, obj, &out)
		if status == C.TCL_OK {
			v.SetUint(uint64(out))
		}
	case reflect.String:
		var n C.int
		out := C.Tcl_GetStringFromObj(obj, &n)
		v.SetString(C.GoStringN(out, n))
	case reflect.Float32, reflect.Float64:
		var out C.double
		status = C.Tcl_GetDoubleFromObj(ir.C, obj, &out)
		if status == C.TCL_OK {
			v.SetFloat(float64(out))
		}
	case reflect.Bool:
		var out C.int
		status = C.Tcl_GetBooleanFromObj(ir.C, obj, &out)
		if status == C.TCL_OK {
			v.SetBool(out == 1)
		}
	default:
		return fmt.Errorf("gothic: cannot convert TCL object to Go type: %s", v.Type())
	}

	if status != C.TCL_OK {
		return errors.New(C.GoString(C.Tcl_GetStringResult(ir.C)))
	}
	return nil
}

type Interpreter struct {
	ir   *interpreter
	lock *sync.Mutex
}

/*
创建一个新的Tcl解析器
可以传入一个脚本内容进行预先执行
*/
func NewInterpreter(initscript string) (ir *Interpreter, err error) {

	ir = new(Interpreter)
	ir.lock = &sync.Mutex{}
	ir.ir, err = new_interpreter()
	if err != nil {
		return nil, err
	}
	if len(initscript) > 0 {
		err = ir.ir.eval([]byte(initscript))
		if err != nil {
			return nil, err
		}
	}

	return ir, nil
}

func appInitProc(*C.Tcl_Interp) int {
	return 0
}

// Queue script for evaluation and wait for its completion. This function uses
// printf-like formatting style, except that all the formatting tags are
// enclosed within %{}. The reason for this is because tcl/tk uses %-based
// formatting tags for its own purposes. Also it provides several extensions
// like named and positional arguments. When no arguments are provided it will
// evaluate the format string as-is.
//
// Another important difference between fmt.Sprintf and Eval is that when using
// Eval, invalid format/arguments combination results in an error, while
// fmt.Sprintf simply ignores misconfiguration. All the formatter generated
// errors go through ErrorFilter, just like any other tcl error.
//
// The syntax for formatting tags is:
//
//	%{[<abbrev>[<format>]]}
//
// Where:
//
//	<abbrev> could be a number of the function argument (starting from 0) or a
//	         name of the key in the provided gothic.ArgMap argument. It can
//	         also be empty, in this case it uses internal counter, takes the
//	         corresponding argument and increments that counter.
//
//	<format> Is the fmt.Sprintf format specifier, passed directly to
//	         fmt.Sprintf as is (except for %q, see additional notes).
//
// Additional notes:
//
//  1. Formatter is extended to do TCL-specific quoting on %q format specifier.
//  2. Named abbrev is only allowed when there is one argument and the type of
//     this argument is gothic.ArgMap.
//
// Examples:
//  1. gothic.Eval("%{0} = %{1} + %{1}", 10, 5)
//     "10 = 5 + 5"
//  2. gothic.Eval("%{} = %{%d} + %{1}", 20, 10)
//     "20 = 10 + 10"
//  3. gothic.Eval("%{0%.2f} and %{%.2f}", 3.1415)
//     "3.14 and 3.14"
//  4. gothic.Eval("[myfunction %{arg1} %{arg2}]", gothic.ArgMap{
//     "arg1": 5,
//     "arg2": 10,
//     })
//     "[myfunction 5 10]"
//  5. gothic.Eval("%{%q}", "[command $variable]")
//     `"\[command \$variable\]"`
func (ir *Interpreter) Eval(format string, args ...interface{}) error {
	ir.lock.Lock()
	defer ir.lock.Unlock()
	ir.ir.cmdbuf.Reset()
	if err := sprintf(&ir.ir.cmdbuf, format, args...); err != nil {
		return err
	}
	return ir.ir.eval(ir.ir.cmdbuf.Bytes())
}

// Works the same way as Eval("%{}", byte_slice), but avoids unnecessary
// buffering.
func (ir *Interpreter) EvalBytes(s []byte) error {
	ir.lock.Lock()
	defer ir.lock.Unlock()
	return ir.ir.eval(s)

}

// Works exactly as Eval with exception that it writes the result of executed
// code into `out`.
func (ir *Interpreter) EvalAs(out interface{}, format string, args ...interface{}) error {
	ir.lock.Lock()
	defer ir.lock.Unlock()
	ir.ir.cmdbuf.Reset()
	if err := sprintf(&ir.ir.cmdbuf, format, args...); err != nil {
		return err
	}
	return ir.ir.eval_as(out, ir.ir.cmdbuf.Bytes())
}

// Shortcut for `var i int; err := EvalAs(&i, format, args...)`
func (ir *Interpreter) EvalAsInt(format string, args ...interface{}) (int, error) {
	var as int
	err := ir.EvalAs(&as, format, args...)
	return as, err
}

// Shortcut for `var s string; err := EvalAs(&s, format, args...)`
func (ir *Interpreter) EvalAsString(format string, args ...interface{}) (string, error) {
	var as string
	err := ir.EvalAs(&as, format, args...)
	return as, err
}

// Shortcut for `var f float64; err := EvalAs(&f, format, args...)`
func (ir *Interpreter) EvalAsFloat(format string, args ...interface{}) (float64, error) {
	var as float64
	err := ir.EvalAs(&as, format, args...)
	return as, err
}

// Shortcut for `var b bool; err := EvalAs(&b, format, args...)`
func (ir *Interpreter) EvalAsBool(format string, args ...interface{}) (bool, error) {
	var as bool
	err := ir.EvalAs(&as, format, args...)
	return as, err
}
