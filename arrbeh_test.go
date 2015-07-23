package qsim

import (
	"testing"
)

func TestShortestQueueArrBeh(t *testing.T) {
	t.Parallel()
	var q *Queue
	var queues []*Queue
	var p *Processor
	var procs []*Processor
	var j *Job
	var ab ArrBeh
	var ass Assignment
	var i, jobsQueued, maxQueueLength int

	queues = make([]*Queue, 3)
	for i = 0; i < 3; i++ {
		queues[i] = NewQueue()
	}
	procs = make([]*Processor, 3)
	for i = 0; i < 3; i++ {
		procs[i] = NewProcessor(simplePtg)
	}
	ab = NewShortestQueueArrBeh(queues, procs)

	// Test that idle Processors get assigned the next Job immediately
	for i = 0; i < 3; i++ {
		ass = ab.Assign(NewJob(0))
		if ass.Type != "Processor" {
			t.Log("Assign returned Assignment with wrong type. Expected 'Processor' but got", ass.Type)
			t.Fail()
		}
	}
	for _, p = range procs {
		if p.IsIdle() {
			t.Log("A Processor is still idle after enough Jobs were assigned for them all")
			t.Fail()
		}
	}
	for _, q = range queues {
		if q.Length() != 0 {
			t.Log("A Job was queued when a Processor should've been idle")
			t.Fail()
		}
	}

	// Test that an arrival when all Processors are busy gets placed in a Queue
	ass = ab.Assign(NewJob(0))
	if ass.Type != "Queue" {
		t.Log("Assign returned Assignment with wrong type. Expected 'Queue' but got", ass.Type)
		t.Fail()
	}
	jobsQueued = 0
	for _, q = range queues {
		jobsQueued += q.Length()
	}
	if jobsQueued != 1 {
		t.Log("A Job arriving to busy Processors was not placed in a Queue")
		t.Fail()
	}

	// Test a few more arrivals
	for i = 0; i < 9; i++ {
		ab.Assign(NewJob(0))
	}
	jobsQueued = 0
	maxQueueLength = 0
	for _, q = range queues {
		jobsQueued += q.Length()
		if q.Length() > maxQueueLength {
			maxQueueLength = q.Length()
		}
	}
	if jobsQueued != 10 {
		t.Log("A Job arriving to busy Processors was not placed in a Queue")
		t.Fail()
	}
	if maxQueueLength != 4 {
		t.Log("A Job arrived and wasn't placed in the shortest Queue")
		t.Fail()
	}

	// Make sure that the correct behavior is followed after a couple Jobs
	// finish.
	procs[0].Finish()
	j, _ = queues[0].Shift()
	procs[0].Start(j)
	procs[1].Finish()
	j, _ = queues[1].Shift()
	procs[1].Start(j)
	ab.Assign(NewJob(0))
	jobsQueued = 0
	for _, q = range queues {
		jobsQueued += q.Length()
	}
	if jobsQueued != 9 {
		t.Log("A Job was queued incorrectly after other Jobs had finished")
		t.Fail()
	}

	// Clear out all the queues and a make a Processor idle, then make sure the
	// next Job gets assigned to that Processor.
	for i = 0; i < 9; i++ {
		procs[i%3].Finish()
		j, _ = queues[i%3].Shift()
		procs[i%3].Start(j)
	}
	procs[1].Finish()
	ab.Assign(NewJob(0))
	if procs[1].IsIdle() {
		t.Log("A Job arriving after the Queues were cleared out did not get assigned directly to a Processor")
		t.Fail()
	}
}

func TestShortestQueueArrBehBeforeAssign(t *testing.T) {
	t.Parallel()
	var queues []*Queue
	var procs []*Processor
	var j, receivedJob *Job
	var ab ArrBeh
	var receivedArrBeh ArrBeh
	var i int

	queues = make([]*Queue, 3)
	for i = 0; i < 3; i++ {
		queues[i] = NewQueue()
	}
	procs = make([]*Processor, 3)
	for i = 0; i < 3; i++ {
		procs[i] = NewProcessor(simplePtg)
	}
	ab = NewShortestQueueArrBeh(queues, procs)

	cbBeforeAssign := func(cbArrBeh ArrBeh, cbJob *Job) *Assignment {
		receivedArrBeh = cbArrBeh
		receivedJob = cbJob
		return nil
	}
	ab.BeforeAssign(cbBeforeAssign)

	j = NewJob(0)
	ab.Assign(j)

	if receivedArrBeh != ab {
		t.Log("BeforeAssign callback ran with wrong ArrBeh or didn't run")
		t.Fail()
	}
	if receivedJob != j {
		t.Log("BeforeAssign callback ran with wrong Job or didn't run")
		t.Fail()
	}
}

// Makes sure that BeforeAssign callbacks can override the ArrBeh's assignment logic.
func TestShortestQueueArrBehBeforeAssignWithOverride(t *testing.T) {
	t.Parallel()
	var queues []*Queue
	var procs []*Processor
	var j, receivedJob *Job
	var ab ArrBeh
	var i int

	queues = make([]*Queue, 3)
	for i = 0; i < 3; i++ {
		queues[i] = NewQueue()
		queues[i].QueueId = i
	}
	procs = make([]*Processor, 3)
	for i = 0; i < 3; i++ {
		procs[i] = NewProcessor(simplePtg)
		procs[i].ProcessorId = i
	}
	ab = NewShortestQueueArrBeh(queues, procs)

	// We create multiple BeforeAssign callbacks to test that the last one to
	// return a non-nil Assignment pointer takes precedence.
	cbBeforeAssignNil := func(cbArrBeh ArrBeh, cbJob *Job) *Assignment { return nil }
	cbBeforeAssignQueue0 := func(cbArrBeh ArrBeh, cbJob *Job) *Assignment {
		return &Assignment{
			Type:  "Queue",
			Queue: queues[0],
		}
	}
	cbBeforeAssignProcessor1 := func(cbArrBeh ArrBeh, cbJob *Job) *Assignment {
		return &Assignment{
			Type:      "Processor",
			Processor: procs[1],
		}
	}
	ab.BeforeAssign(cbBeforeAssignNil)
	ab.BeforeAssign(cbBeforeAssignProcessor1)
	ab.BeforeAssign(cbBeforeAssignQueue0)
	ab.BeforeAssign(cbBeforeAssignNil)

	j = NewJob(0)
	ab.Assign(j)

	if receivedJob, _ = queues[0].Shift(); receivedJob != j {
		t.Log("Job was not assigned to Queue specified by BeforeAssign callback")
		t.Fail()
	}
}

func TestShortestQueueArrBehAfterAssign(t *testing.T) {
	t.Parallel()
	var queues []*Queue
	var procs []*Processor
	var j, receivedJob *Job
	var ab ArrBeh
	var receivedArrBeh ArrBeh
	var receivedAssignment Assignment
	var i int

	queues = make([]*Queue, 3)
	for i = 0; i < 3; i++ {
		queues[i] = NewQueue()
	}
	procs = make([]*Processor, 3)
	for i = 0; i < 3; i++ {
		procs[i] = NewProcessor(simplePtg)
	}
	ab = NewShortestQueueArrBeh(queues, procs)

	cbAfterAssign := func(cbArrBeh ArrBeh, cbJob *Job, cbAssignment Assignment) {
		receivedArrBeh = cbArrBeh
		receivedJob = cbJob
		receivedAssignment = cbAssignment
	}
	ab.AfterAssign(cbAfterAssign)

	j = NewJob(0)
	ab.Assign(j)

	if receivedArrBeh != ab {
		t.Log("AfterAssign callback ran with wrong ArrBeh or didn't run")
		t.Fail()
	}
	if receivedJob != j {
		t.Log("AfterAssign callback ran with wrong Job or didn't run")
		t.Fail()
	}
	if receivedAssignment.Type != "Processor" {
		t.Log("Expected Assignment Type 'Processor' but got", receivedAssignment.Type)
		t.Fail()
	}
	if receivedAssignment.Processor == nil {
		t.Log("Assignment Type was 'Processor' but Processor = nil")
	}

	// Now test what happens when the assignment is to a Queue rather than a Processor
	ab.Assign(NewJob(0))
	ab.Assign(NewJob(0))
	j = NewJob(0)
	ab.Assign(j)

	if receivedArrBeh != ab {
		t.Log("AfterAssign callback ran with wrong ArrBeh or didn't run")
		t.Fail()
	}
	if receivedJob != j {
		t.Log("AfterAssign callback ran with wrong Job or didn't run")
		t.Fail()
	}
	if receivedAssignment.Type != "Queue" {
		t.Log("Expected Assignment Type 'Queue' but got", receivedAssignment.Type)
		t.Fail()
	}
	if receivedAssignment.Queue == nil {
		t.Log("Assignment Type was 'Queue' but Queue = nil")
	}
}
