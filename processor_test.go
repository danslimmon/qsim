package qsim

import (
	"testing"
)

// Tests the starting of a job
func TestProcessorStart(t *testing.T) {
	var proc *Processor
	var j *Job
	var procTime int

	ptg := func(j *Job) int {
		return 293
	}

	proc = NewProcessor()
	proc.SetProcTimeGenerator(ptg)
	j = NewJob()

	procTime = proc.Start(j)
	if procTime != 293 {
		t.Log("Expected processing time of 293 but got", procTime)
		t.Fail()
	}
}
