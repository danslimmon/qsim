package qsim

import (
	"testing"
)

// A simple ProcTimeGenerator function that returns a constant
func simplePtg(j *Job) int {
	return 293
}

// Tests the starting of a job
func TestProcessorStart(t *testing.T) {
	var proc *Processor
	var j0, j1 *Job
	var procTime int
	var err error

	proc = NewProcessor()
	proc.SetProcTimeGenerator(simplePtg)
	j0 = NewJob()

	procTime, err = proc.Start(j0)
	if procTime != 293 {
		t.Log("Expected processing time of 293 but got", procTime)
		t.Fail()
	}
	if err != nil {
		t.Log("Got unexpected error from proc.Start:", err)
		t.Fail()
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
		t.Fail()
	}
}

// Tests the finishing of a job
func TestProcessorFinish(t *testing.T) {
	var proc *Processor
	var j *Job

	proc = NewProcessor()
	proc.SetProcTimeGenerator(simplePtg)
	j = NewJob()
	proc.Start(j)
	if j != proc.Finish() {
		t.Log("Expected to get back from proc.Finish the job that was processing")
		t.Fail()
	}

	if nil != proc.Finish() {
		t.Log("proc.Finish on an idle job didn't return nil as expected")
		t.Fail()
	}
}
