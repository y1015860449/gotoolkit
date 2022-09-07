package coPool

import (
	"fmt"
	"testing"
	"time"
)

type tmpStruct struct {
	index int32
	data  string
}

func TestCoPool(t *testing.T) {
	p := NewPool(10, 2, 2*time.Second)
	testHandler := func(param interface{}) error {
		testData := param.(*tmpStruct)
		t.Logf("testHandler data(%v)", testData)
		return nil
	}

	testHandler22222 := func(param interface{}) error {
		testData := param.(*tmpStruct)
		t.Logf("testHandler22222 data(%v)", testData)
		return nil
	}

	for i := 0; i < 10000; i++ {
		st := &tmpStruct{
			index: int32(i),
			data:  fmt.Sprintf("test pool index %d", i),
		}
		if i%2 == 1 {
			if err := p.AddTask(NewTask(st, testHandler)); err != nil {
				t.Errorf("Submit err(%+v) param(%v)", err, st)
			}
		} else {
			if err := p.AddTask(NewTask(st, testHandler22222)); err != nil {
				t.Errorf("Submit2 err(%+v) param(%v)", err, st)
			}
		}

	}
	time.Sleep(6 * time.Second)
	p.Close()
	time.Sleep(4 * time.Second)
}
