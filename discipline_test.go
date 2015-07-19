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
		queues[i].QueueId = i
		procs = append(procs, NewProcessor(simplePtg))
		procs[i].ProcessorId = i

		// Make sure that Processors only pull work from their own queues.
		procs[i].BeforeStart(func(p *Processor, j *Job) {
			var queueId int
			queueId = j.IntAttrs["queue_id"]
			if queueId != p.ProcessorId {
				t.Log("Processor", p.ProcessorId, "pulled a Job from Queue", queueId)
				t.Fail()
			}
		})
	}
	NewOneToOneFIFODiscipline(queues, procs)

	for i = 0; i < 3; i++ {
		j = NewJob(0)
		j.IntAttrs["queue_id"] = i
		procs[i].Start(j)
	}
	for i = 0; i < 6; i++ {
		j = NewJob(0)
		j.IntAttrs["queue_id"] = i % 3
		queues[i%3].Append(j)
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
}
