package qsim

import (
	"testing"
)

// Tests a dead-simple Arrival Process
func TestConstantArrProc(t *testing.T) {
	t.Parallel()
	var ap ArrProc
	var j *Job
	var i, time, interval int

	ap = NewConstantArrProc(72)
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

// Tests the BeforeArrive callback on ConstantArrProc.
func TestConstantArrProcBeforeArrive(t *testing.T) {
	t.Parallel()
	var ap, receivedArrProc ArrProc

	cbBeforeArrive := func(cbArrProc ArrProc) {
		receivedArrProc = cbArrProc
	}

	ap = NewConstantArrProc(72)
	ap.BeforeArrive(cbBeforeArrive)
	ap.Arrive()
	if ap != receivedArrProc {
		t.Log("BeforeArrive ran with wrong ArrProc or didn't run")
		t.Fail()
	}
}

// Tests the AfterArrive callback on ConstantArrProc.
func TestConstantArrProcAfterArrive(t *testing.T) {
	t.Parallel()
	var ap, receivedArrProc ArrProc
	var j, receivedJob *Job
	var interval, receivedInterval int

	cbAfterArrive := func(cbArrProc ArrProc, cbJob *Job, cbInterval int) {
		receivedArrProc = cbArrProc
		receivedJob = cbJob
		receivedInterval = cbInterval
	}

	ap = NewConstantArrProc(72)
	ap.AfterArrive(cbAfterArrive)
	j, interval = ap.Arrive()
	if ap != receivedArrProc {
		t.Log("AfterArrive ran with wrong ArrProc or didn't run")
		t.Fail()
	}
	if j != receivedJob {
		t.Log("AfterArrive ran with wrong Job or didn't run")
		t.Fail()
	}
	if interval != receivedInterval {
		t.Log("AfterArrive ran with wrong interval or didn't run")
		t.Fail()
	}
}
