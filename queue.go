package qsim

// A Queue holds Jobs until they're ready for processing.
type Queue struct {
	// The Jobs currently in the queue
	Jobs []*Job
	// A unique identifier for the Queue. Useful for debugging, as it
	// will be printed in debug output for events involving the Queue.
	// The implementor must set this value if it's going to be used –
	// otherwise it will be 0 (and thus not unique)
	QueueId int
	// The maximum length of the Queue. The default, -1, allows
	// arbitrarily many events to be appended. At any other value
	// (including 0), Append() calls will still succeed but the Job
	// will be discarded instead of appended.
	//
	// MaxLength may be raised during the course of a simulation. It
	// may also be lowered, with the effect that the Jobs currently in
	// the Queue will remain there, but new Jobs won't be appended
	// until the Queue's length is back under MaxLength.
	MaxLength int

	// Callback lists
	cbBeforeAppend []func(q *Queue, j *Job)
	cbAfterAppend  []func(q *Queue, j *Job)
	cbBeforeShift  []func(q *Queue, j *Job)
	cbAfterShift   []func(q *Queue, j *Job)
	cbBeforeRemove []func(q *Queue, j *Job)
	cbAfterRemove  []func(q *Queue, j *Job)
}

// Append adds a Job to the tail of the queue.
func (q *Queue) Append(j *Job) {
	q.beforeAppend(j)
	if q.MaxLength == -1 || q.Length() < q.MaxLength {
		q.Jobs = append(q.Jobs, j)
		q.afterAppend(j)
	} else {
		q.afterAppend(nil)
	}
}

// Length returns the current number of jobs in the queue.
func (q *Queue) Length() int {
	return len(q.Jobs)
}

// Shift removes a Job from the head of the queue.
//
// It returns the Job that was removed, as well as the number of Jobs
// still left in the queue after shifting. When Shift is called on an
// empty queue, j will be nil. So an appropriate use of Shift looks like
// this:
//
//  j, nrem := q.Shift()
//  if j != nil {
//      // Do something with j
//  }
func (q *Queue) Shift() (j *Job, nrem int) {
	if len(q.Jobs) == 0 {
		q.beforeShift(nil)
		q.afterShift(nil)
		return nil, 0
	}
	j = q.Jobs[0]
	q.beforeShift(j)
	q.Jobs = q.Jobs[1:]
	q.afterShift(j)
	return j, len(q.Jobs)
}

// Remove removes a particular Job (identified by JobId property) from the queue.
//
// It returns the Job that was removed, as well as the number of Jobs
// still left in the queue after removing. When Remove is passed a Job that
// is absent from the queue, its returned job will be nil. So an appropriate
// use of Remove() looks like this:
//
//  j, nrem := q.Remove(j)
//  if j != nil {
//      // Do something with j
//  }
func (q *Queue) Remove(jToRemove *Job) (jRet *Job, nrem int) {
	for i, j := range q.Jobs {
		if j.JobId == jToRemove.JobId {
			q.beforeRemove(jToRemove)
			q.Jobs = append(q.Jobs[:i], q.Jobs[i+1:]...)
			q.afterRemove(jToRemove)
			return jToRemove, q.Length()
		}
	}
	return nil, q.Length()
}

// BeforeAppend adds a callback to be run immediately before a Job is
// appended to the queue.
//
// The callback will be passed the queue itself and the job that's about
// to be appended.
func (q *Queue) BeforeAppend(f func(q *Queue, j *Job)) {
	q.cbBeforeAppend = append(q.cbBeforeAppend, f)
}
func (q *Queue) beforeAppend(j *Job) {
	for _, cb := range q.cbBeforeAppend {
		cb(q, j)
	}
}

// AfterAppend adds a callback to be run immediately after a Job is
// appended to the queue.
//
// The callback will be passed the queue itself and the job that was just
// appended. If Append is called when the Queue's length is not below
// MaxLength, AfterAppend callbacks will still run but they'll be passed
// j=nil.
func (q *Queue) AfterAppend(f func(q *Queue, j *Job)) {
	q.cbAfterAppend = append(q.cbAfterAppend, f)
}
func (q *Queue) afterAppend(j *Job) {
	for _, cb := range q.cbAfterAppend {
		cb(q, j)
	}
}

// BeforeShift adds a callback to be run immediately before a Job is
// shifted out of the queue.
//
// The callback will be passed the queue itself and the job that's about
// to be shifted. If Shift is called on an empty queue, this callback
// will run but j will be nil.
func (q *Queue) BeforeShift(f func(q *Queue, j *Job)) {
	q.cbBeforeShift = append(q.cbBeforeShift, f)
}
func (q *Queue) beforeShift(j *Job) {
	for _, cb := range q.cbBeforeShift {
		cb(q, j)
	}
}

// AfterShift adds a callback to be run immediately after a Job is
// shifted out of the queue.
//
// The callback will be passed the queue itself and the job that was
// just shifted. If Shift is called on an empty queue, this callback
// will run but j will be nil.
func (q *Queue) AfterShift(f func(q *Queue, j *Job)) {
	q.cbAfterShift = append(q.cbAfterShift, f)
}
func (q *Queue) afterShift(j *Job) {
	for _, cb := range q.cbAfterShift {
		cb(q, j)
	}
}

// BeforeRemove adds a callback to be run immediately before a Job is
// removed from the queue (with Remove()).
//
// The callback will be passed the queue itself and the job that's about
// to be removed. If Remove is called on an empty queue, this callback
// will not run.
func (q *Queue) BeforeRemove(f func(q *Queue, j *Job)) {
	q.cbBeforeRemove = append(q.cbBeforeRemove, f)
}
func (q *Queue) beforeRemove(j *Job) {
	for _, cb := range q.cbBeforeRemove {
		cb(q, j)
	}
}

// AfterRemove adds a callback to be run immediately after a Job is
// removed from the queue (with Remove()).
//
// The callback will be passed the queue itself and the job that was
// just shifted. If Remove is called on an empty queue, this callback
// will not run.
func (q *Queue) AfterRemove(f func(q *Queue, j *Job)) {
	q.cbAfterRemove = append(q.cbAfterRemove, f)
}
func (q *Queue) afterRemove(j *Job) {
	for _, cb := range q.cbAfterRemove {
		cb(q, j)
	}
}

// NewQueue creates an empty Queue.
func NewQueue() (q *Queue) {
	q = new(Queue)
	q.Jobs = make([]*Job, 0)
	q.MaxLength = -1
	return q
}
