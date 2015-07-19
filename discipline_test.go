package qsim

import (
	"testing"
)

// Tests the behavior of a OneToOneFIFODiscipline
func TestOneToOneFIFODiscipline(t *testing.T) {
	t.Parallel()
	var queues []*Queue
	var procs []*Processor
	var j *Job
	var i int

	for i = 0; i < 3; i++ {
		queues = append(queues, NewQueue())
		procs = append(procs, NewProcessor(simplePtg))
	}
	NewOneToOneFIFODiscipline(queues, procs)

	for i = 0; i < 3; i++ {
		procs[i].Start(NewJob(0))
	}
	for i = 0; i < 6; i++ {
		queues[i%3].Append(NewJob(0))
	}

	j = queues[2].Jobs[0]
	procs[2].Finish()
	if procs[2].CurrentJob != j {
		t.Log("Processor should have been assigned the oldest job in its queue, but instead was assigned ", procs[2].CurrentJob)
		t.Fail()
	}
	if queues[2].Length() != 1 {
		t.Log("A Job should've been shifted out of the Queue, but Queue is still the same length as before")
		t.Fail()
	}

	// Make sure that Processors only pull work from their own Queue
	procs[2].Finish()
	procs[2].Finish()
	if !procs[2].IsIdle() {
		t.Log("Processor is not idle, but it shouldn't have had any more work in its queue")
		t.Fail()
	}
	if queues[2].Length() != 0 {
		t.Log("Processor finished all work in its Queue, but the Queue isn't empty")
		t.Fail()
	}
}
