package qsim

import (
	"testing"
)

// Tests a dead-simple Arrival Process
func TestConstantArrProc(t *testing.T) {
	t.Parallel()
	var ap ArrProc
	var j *Job
	var jobs []*Job
	var i, time, interval int

	ap = NewConstantArrProc(72)
	for i = 0; i < 10; i++ {
		jobs, interval = ap.Arrive(0)
		if len(jobs) != 1 {
			t.Logf("ConstantArrProc.Arrive returned %d jobs but we expected exactly 1", len(jobs))
			t.Fail()
		}
		j = jobs[0]
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
	ap.Arrive(0)
	if ap != receivedArrProc {
		t.Log("BeforeArrive ran with wrong ArrProc or didn't run")
		t.Fail()
	}
}

// Tests the AfterArrive callback on ConstantArrProc.
func TestConstantArrProcAfterArrive(t *testing.T) {
	t.Parallel()
	var ap, receivedArrProc ArrProc
	var j *Job
	var jobs, receivedJobs []*Job
	var interval, receivedInterval int

	cbAfterArrive := func(cbArrProc ArrProc, cbJobs []*Job, cbInterval int) {
		receivedArrProc = cbArrProc
		receivedJobs = cbJobs
		receivedInterval = cbInterval
	}

	ap = NewConstantArrProc(72)
	ap.AfterArrive(cbAfterArrive)
	jobs, interval = ap.Arrive(0)
	if len(jobs) != 1 {
		t.Logf("Arrive returned %d jobs but we expected exactly 1", len(jobs))
		t.Fail()
	}
	j = jobs[0]
	if ap != receivedArrProc {
		t.Log("AfterArrive ran with wrong ArrProc or didn't run")
		t.Fail()
	}
	if len(receivedJobs) != 1 {
		t.Logf("AfterArrive ran with %d jobs but we expected exactly 1", len(receivedJobs))
		t.Fail()
	}
	if j != receivedJobs[0] {
		t.Log("AfterArrive ran with wrong Job or didn't run")
		t.Fail()
	}
	if interval != receivedInterval {
		t.Log("AfterArrive ran with wrong interval or didn't run")
		t.Fail()
	}
}

// Tests a dead-simple Arrival Process
func TestPoissonArrProc(t *testing.T) {
	t.Parallel()
	var ap ArrProc
	var j *Job
	var jobs []*Job
	var i, time, interval int

	// Poisson arrival process with a mean arrival interval of 1000
	// ticks.
	ap = NewPoissonArrProc(1000)
	for i = 0; i < 1000; i++ {
		jobs, interval = ap.Arrive(0)
		if len(jobs) != 1 {
			t.Logf("Arrive returned with %d jobs but we expected exactly 1", len(jobs))
			t.Fail()
		}
		j = jobs[0]
		if j == nil {
			t.Log("PoissonArrProc.Arrive returned nil job")
			t.Fail()
		}
		time += interval
	}
	if time < 800*1000 || time > 1200*1000 {
		// The probability of the mean being this far away from 1000 is extremely low, so
		// this test should be fine.
		t.Log("Average arrival interval from PoissonArrProc is too far from 1000: got", time/1000)
		t.Fail()
	}
}
