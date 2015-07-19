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
	if len(q.Jobs) != 20 {
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
