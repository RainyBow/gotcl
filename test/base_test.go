package test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/RainyBow/gotcl"
	"github.com/shirou/gopsutil/process"
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
		ir.Free()
		for i := 0; i < 1000; i++ {
			log.Printf("i=%d", i)
			time.Sleep(10 * time.Second)
		}
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

func TestCase3(t *testing.T) {
	ir, err := gotcl.NewInterpreter(`
	puts $auto_path
	`)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < 10; i++ {
		log.Printf("i0=%d", i)
		time.Sleep(10 * time.Second)
	}
	ir.Free()
	for i := 0; i < 1000; i++ {
		log.Printf("i=%d", i)
		time.Sleep(10 * time.Second)
	}

}

func TestCase4(t *testing.T) {
	for i := 0; i < 10000; i++ {
		log.Printf("i=%d: memory:%s", i, getCurrentMemory())
		if ir, err := gotcl.NewInterpreter(`
		lappend auto_path /opt/Spirent_TestCenter_5.29/Spirent_TestCenter_Application_Linux
		`); err == nil {
			if ver, err := ir.EvalAsString("package require stc"); err == nil {
				log.Printf("i=%d:ok ver:%s", i, ver)
				ir.Free()
				time.Sleep(2 * time.Second)
				ir = nil
				continue
			}
		}
		t.Fail()
		break

	}
	fmt.Printf("end: memory:%s", getCurrentMemory())

}

func getCurrentMemory() string {
	pid := os.Getpid()
	if proc, err := process.NewProcess(int32(pid)); proc != nil && err == nil {
		if mem_info, err := proc.MemoryInfo(); err == nil && mem_info != nil {
			return mem_info.String()
		}
	}
	return ""
}
