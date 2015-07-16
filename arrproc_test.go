package qsim

import (
	"testing"
)

func TestConstantArrProc(t *testing.T) {
	t.Parallel()
	var ap ArrProc
	var j *Job
	var i, time, interval int

	ap = &ConstantArrProc{72}
	for i = 0; i < 10; i++ {
		j, interval = ap.Arrive()
		if j == nil {
			t.Log("ConstantArrProc.Arrive returned nil job")
			t.Fail()
		}
		time += interval
	}
	if time != 720 {
		t.Log("Expected", 720, "ticks to elapse from ConstantArrProc arrivals but got", time)
		t.Fail()
	}
}
