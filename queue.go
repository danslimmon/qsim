package qsim

// A Queue holds Jobs until they're ready for processing.
type Queue struct {
	// The Jobs currently in the queue
	Jobs []*Job

	// Callback lists
	cbBeforeAppend []func(q *Queue, j *Job)
	cbAfterAppend  []func(q *Queue, j *Job)
	cbBeforeShift  []func(q *Queue, j *Job)
	cbAfterShift   []func(q *Queue, j *Job)
}

// Append adds a Job to the tail of the queue.
func (q *Queue) Append(j *Job) {
	q.beforeAppend(j)
	q.Jobs = append(q.Jobs, j)
	q.afterAppend(j)
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
// appended.
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

// NewQueue creates an empty Queue.
func NewQueue() (q *Queue) {
	q = new(Queue)
	q.Jobs = make([]*Job, 0)
	return q
}
