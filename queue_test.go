package qsim

import (
	"testing"
)

// Tests serial queue insertion
func TestQueueAppend(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	for i := 0; i < 20; i++ {
		j = NewJob(0)
		j.IntAttrs["i"] = i
		q.Append(j)
	}

	if q.Jobs[0].IntAttrs["i"] != 0 {
		t.Log("Zeroeth element not found at index 0")
		t.Fail()
	}
	if q.Jobs[19].IntAttrs["i"] != 19 {
		t.Log("Nineteenth element not found at index 19")
		t.Fail()
	}
	if q.Length() != 20 {
		t.Log("Wrong number of Jobs made it into Queue")
		t.Fail()
	}
}

// Tests the behavior of shifting items out of the queue.
func TestQueueShift(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j0, j1, j *Job
	var nrem int
	q = NewQueue()

	j0 = NewJob(0)
	j0.IntAttrs["i"] = 0
	q.Append(j0)

	j1 = NewJob(0)
	j1.IntAttrs["i"] = 1
	q.Append(j1)

	j, nrem = q.Shift()
	if j.IntAttrs["i"] != 0 {
		t.Log("Zeroeth element was not the first to be shifted")
		t.Fail()
	}
	if nrem != 1 {
		t.Log("Expected the remaining number of queued jobs to equal 1")
		t.Fail()
	}

	j, nrem = q.Shift()
	if j.IntAttrs["i"] != 1 {
		t.Log("Last element was not the last to be shifted")
		t.Fail()
	}
	if nrem != 0 {
		t.Log("Expected the remaining number of queued jobs to equal 0")
		t.Fail()
	}

	j, nrem = q.Shift()
	if j != nil {
		t.Log("Calling Shift on an empty queue should return j = nil")
		t.Fail()
	}
}

// Tests the behavior of removing particular items from the queue.
func TestQueueRemove(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j0, j1, j2, j *Job
	var nrem int
	q = NewQueue()

	j0 = NewJob(0)
	j0.IntAttrs["i"] = 0
	q.Append(j0)

	j1 = NewJob(0)
	j1.IntAttrs["i"] = 1
	q.Append(j1)

	j, nrem = q.Remove(j1)
	if j.IntAttrs["i"] != 1 {
		t.Log("Incorrect job removed")
		t.Fail()
	}
	if nrem != 1 {
		t.Log("Expected the remaining number of queued jobs to equal 1")
		t.Fail()
	}

	j, nrem = q.Remove(j0)
	if j.IntAttrs["i"] != 0 {
		t.Log("Incorrect job removed")
		t.Fail()
	}
	if nrem != 0 {
		t.Log("Expected the remaining number of queued jobs to equal 0")
		t.Fail()
	}

	j2 = NewJob(0)
	j, nrem = q.Remove(j2)
	if j != nil {
		t.Log("Calling Remove on a non-queued job should return j = nil")
		t.Fail()
	}
}

// Tests the BeforeAppend callback.
func TestQueueBeforeAppend(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		counter = j.IntAttrs["i"]
		qLen = len(q.Jobs)
	}
	q.BeforeAppend(cb)

	j = NewJob(0)
	j.IntAttrs["i"] = 0
	q.Append(j)
	if counter != 0 {
		t.Log("Expected BeforeAppend callback to set counter=0")
		t.Fail()
	}
	if qLen != 0 {
		t.Log("Expected BeforeAppend callback to set qLen=0")
		t.Fail()
	}

	j = NewJob(0)
	j.IntAttrs["i"] = 1
	q.Append(j)
	if counter != 1 {
		t.Log("Expected BeforeAppend callback to set counter=1")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected BeforeAppend callback to set qLen=1")
		t.Fail()
	}
}

// Tests the AfterAppend callback.
func TestQueueAfterAppend(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		counter = j.IntAttrs["i"]
		qLen = len(q.Jobs)
	}
	q.AfterAppend(cb)

	j = NewJob(0)
	j.IntAttrs["i"] = 0
	q.Append(j)
	if counter != 0 {
		t.Log("Expected AfterAppend callback to set counter=0")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected AfterAppend callback to set qLen=1")
		t.Fail()
	}

	j = NewJob(0)
	j.IntAttrs["i"] = 1
	q.Append(j)
	if counter != 1 {
		t.Log("Expected AfterAppend callback to set counter=1")
		t.Fail()
	}
	if qLen != 2 {
		t.Log("Expected AfterAppend callback to set qLen=2")
		t.Fail()
	}
}

// Tests the BeforeShift callback.
func TestQueueBeforeShift(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		if j != nil {
			counter = j.IntAttrs["i"]
		} else {
			counter = -1
		}
		qLen = len(q.Jobs)
	}
	q.BeforeShift(cb)

	j = NewJob(0)
	j.IntAttrs["i"] = 0
	q.Append(j)
	j = NewJob(0)
	j.IntAttrs["i"] = 1
	q.Append(j)

	q.Shift()
	if counter != 0 {
		t.Log("Expected BeforeShift callback to set counter=0")
		t.Fail()
	}
	if qLen != 2 {
		t.Log("Expected BeforeShift callback to set qLen=2")
		t.Fail()
	}

	q.Shift()
	if counter != 1 {
		t.Log("Expected BeforeShift callback to set counter=1")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected BeforeShift callback to set qLen=1")
		t.Fail()
	}

	q.Shift()
	if counter != -1 {
		t.Log("Expected BeforeShift callback to set counter=-1")
		t.Fail()
	}
	if qLen != 0 {
		t.Log("Expected BeforeShift callback to set qLen=0")
		t.Fail()
	}
}

// Tests the AfterShift callback.
func TestQueueAfterShift(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		if j != nil {
			counter = j.IntAttrs["i"]
		} else {
			counter = -1
		}
		qLen = len(q.Jobs)
	}
	q.AfterShift(cb)

	j = NewJob(0)
	j.IntAttrs["i"] = 0
	q.Append(j)
	j = NewJob(0)
	j.IntAttrs["i"] = 1
	q.Append(j)

	q.Shift()
	if counter != 0 {
		t.Log("Expected AfterShift callback to set counter=0")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected AfterShift callback to set qLen=1")
		t.Fail()
	}

	q.Shift()
	if counter != 1 {
		t.Log("Expected AfterShift callback to set counter=1")
		t.Fail()
	}
	if qLen != 0 {
		t.Log("Expected AfterShift callback to set qLen=0")
		t.Fail()
	}

	q.Shift()
	if counter != -1 {
		t.Log("Expected AfterShift callback to set counter=-1")
		t.Fail()
	}
	if qLen != 0 {
		t.Log("Expected AfterShift callback to set qLen=0")
		t.Fail()
	}
}

// Tests the BeforeRemove callback.
func TestQueueBeforeRemove(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j0, j1, j2 *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		if j != nil {
			counter = j.IntAttrs["i"]
		} else {
			counter = -1
		}
		qLen = q.Length()
	}
	q.BeforeRemove(cb)

	j0 = NewJob(0)
	j0.IntAttrs["i"] = 0
	q.Append(j0)
	j1 = NewJob(0)
	j1.IntAttrs["i"] = 1
	q.Append(j1)

	j2 = NewJob(0)
	j2.IntAttrs["i"] = 1

	// BeforeRemove should not run if the job is not present.
	counter = -2
	qLen = -2
	q.Remove(j2)
	if counter != -2 {
		t.Log("Expected BeforeRemove not to run")
		t.Fail()
	}
	if qLen != -2 {
		t.Log("Expected BeforeRemove not to run")
		t.Fail()
	}

	q.Remove(j1)
	if counter != 1 {
		t.Log("Expected BeforeRemove callback to set counter=1")
		t.Fail()
	}
	if qLen != 2 {
		t.Log("Expected BeforeRemove callback to set qLen=2")
		t.Fail()
	}

	q.Remove(j0)
	if counter != 0 {
		t.Log("Expected BeforeRemove callback to set counter=0")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected BeforeRemove callback to set qLen=1")
		t.Fail()
	}

	// This job is no longer present, so BeforeRemove shouldn't run.
	counter = -2
	qLen = -2
	q.Remove(j0)
	if counter != -2 {
		t.Log("Expected BeforeRemove not to run")
		t.Fail()
	}
	if qLen != -2 {
		t.Log("Expected BeforeRemove not to run")
		t.Fail()
	}
}

// Tests the AfterRemove callback.
func TestQueueAfterRemove(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j0, j1, j2 *Job
	q = NewQueue()

	var counter int
	var qLen int
	cb := func(q *Queue, j *Job) {
		if j != nil {
			counter = j.IntAttrs["i"]
		} else {
			counter = -1
		}
		qLen = q.Length()
	}
	q.AfterRemove(cb)

	j0 = NewJob(0)
	j0.IntAttrs["i"] = 0
	q.Append(j0)
	j1 = NewJob(0)
	j1.IntAttrs["i"] = 1
	q.Append(j1)

	j2 = NewJob(0)
	j2.IntAttrs["i"] = 2

	// AfterRemove() shouldn't run on an absent job
	counter = -2
	qLen = -2
	q.Remove(j2)
	if counter != -2 {
		t.Log("Expected AfterRemove callback not to run")
		t.Fail()
	}
	if qLen != -2 {
		t.Log("Expected AfterRemove callback not to run")
		t.Fail()
	}

	q.Remove(j1)
	if counter != 1 {
		t.Log("Expected AfterRemove callback to set counter=1")
		t.Fail()
	}
	if qLen != 1 {
		t.Log("Expected AfterRemove callback to set qLen=1")
		t.Fail()
	}

	q.Remove(j0)
	if counter != 0 {
		t.Log("Expected AfterRemove callback to set counter=0")
		t.Fail()
	}
	if qLen != 0 {
		t.Log("Expected AfterRemove callback to set qLen=0")
		t.Fail()
	}

	// j0 is no longer present, so AfterRemove callback shouldn't run.
	counter = -2
	qLen = -2
	q.Remove(j0)
	if counter != -2 {
		t.Log("Expected AfterRemove callback to set counter=-2")
		t.Fail()
	}
	if qLen != -2 {
		t.Log("Expected AfterRemove callback to set qLen=-2")
		t.Fail()
	}
}

// Tests queue insertion with a MaxLength
func TestQueueAppendWithMaxLength(t *testing.T) {
	t.Parallel()
	var q *Queue
	var i int

	q = NewQueue()
	q.MaxLength = 3

	for i := 0; i < 3; i++ {
		q.Append(NewJob(0))
	}
	if q.Length() != 3 {
		t.Log("Queue with MaxLength didn't accept all the Jobs it should have")
		t.Fail()
	}

	q.Append(NewJob(0))
	if q.Length() > 3 {
		t.Log("Queue grew longer than its MaxLength")
		t.Fail()
	}

	// When we lower the MaxLength of a Queue, we expect it to keep its current
	// contents but not accept new Jobs until it's below that length.
	q.MaxLength = 2
	q.Append(NewJob(0))
	if q.Length() > 3 {
		t.Log("Queue accepted new Job after its MaxLength was lowered")
		t.Fail()
	}
	if q.Length() < 3 {
		t.Log("Queue dropped Jobs when its MaxLength was lowered")
		t.Fail()
	}

	q.Shift()
	q.Shift()
	if q.Length() > 1 {
		t.Log("Failed to shift Jobs out of Queue after its MaxLength was lowered")
		t.Fail()
	}
	q.Append(NewJob(0))
	q.Append(NewJob(0))
	if q.Length() < 2 {
		t.Log("Failed to append Jobs up to Queue's new MaxLength")
		t.Fail()
	}
	if q.Length() > 2 {
		t.Log("Queue accepted Jobs beyond its new MaxLength")
		t.Fail()
	}

	// Now when we raise MaxLength we should be able to append another Job
	q.MaxLength = 3
	q.Append(NewJob(0))
	if q.Length() < 3 {
		t.Log("Failed to append Job after raising Queue's MaxLength")
		t.Fail()
	}

	// If we set MaxLength to -1, we should be able to append as many Jobs
	// as we want.
	q.MaxLength = -1
	for i = 0; i < 30; i++ {
		q.Append(NewJob(0))
	}
	if q.Length() != 33 {
		t.Log("Failed to add a bunch more Jobs after setting Queue to unlimited MaxLength")
		t.Fail()
	}
}

// Tests the behavior of AfterAppend when appending against a MaxLength
func TestQueueAfterAppendWithMaxLength(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j, receivedJob *Job
	q = NewQueue()
	q.MaxLength = 1

	cbAfterAppend := func(cbQueue *Queue, cbJob *Job) {
		receivedJob = cbJob
	}
	q.AfterAppend(cbAfterAppend)

	// Test the normal behavior, before we reach MaxLength
	j = NewJob(0)
	q.Append(j)
	if receivedJob != j {
		t.Log("AfterAppend got wrong Job with MaxLength set; expected", j, "but got", receivedJob)
		t.Fail()
	}

	// If we try to append at MaxLength, the callback should get <nil> as its Job
	q.Append(NewJob(0))
	if receivedJob != nil {
		t.Log("AfterAppend at MaxLength should get passed nil Job, but got", receivedJob)
		t.Fail()
	}
}
