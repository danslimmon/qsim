package qsim

import (
	"testing"
)

// Tests the starting of a job
func TestProcessorStart(t *testing.T) {
	var proc *Processor
	var j0, j1 *Job
	var procTime int
	var err error

	ptg := func(j *Job) int {
		return 293
	}

	proc = NewProcessor()
	proc.SetProcTimeGenerator(ptg)
	j0 = NewJob()

	procTime, err = proc.Start(j0)
	if procTime != 293 {
		t.Log("Expected processing time of 293 but got", procTime)
		t.Fail()
	}
	if err != nil {
		t.Log("Got unexpected error from proc.Start:", err)
	}

	// Make sure we get an error if we try to start a job while the
	// processor is busy.
	j1 = NewJob()
	procTime, err = proc.Start(j1)
	if err == nil {
		t.Log("Expected 'job already in progress error', got no error")
		t.Fail()
	}

	// We should still be able to start that new job as long as we
	// finish the first job:
	proc.Finish()
	procTime, err = proc.Start(j1)
	if procTime != 293 {
		t.Log("Expected processing time of 293 but got", procTime)
		t.Fail()
	}
	if err != nil {
		t.Log("Got unexpected error from proc.Start:", err)
	}
}
