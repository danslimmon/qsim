package qsim

// A Queue holds Jobs until they're ready for processing.
type Queue struct {
	// The Jobs currently in the queue
	Jobs []*Job

	// cbBeforeAppend holds the callbacks to run immediately before appending to
	// the queue.
	cbBeforeAppend []func(q *Queue, j *Job)
}

// Append adds a Job to the tail of the queue.
func (q *Queue) Append(j *Job) {
	q.beforeAppend(j)
	q.Jobs = append(q.Jobs, j)
}

// Shift removes a Job from the head of the queue.
//
// It returns the Job that was removed, as well as the number of Jobs
// still left in the queue after shifting. When Shift is called on an
// empty queue, j will be nil. So an appropriate use of Shift looks like
// this:
//
//  j, nrem := q.Shift()
//	if j != nil {
//		// Do something with j
//  }
func (q *Queue) Shift() (j *Job, nrem int) {
	if len(q.Jobs) == 0 {
		return nil, 0
	}
	j = q.Jobs[0]
	q.Jobs = q.Jobs[1:]
	return j, len(q.Jobs)
}

// OnBeforeAppend adds a callback to be run immediately before a Job is
// appended to the queue.
//
// The callback will be passed the queue itself and the job that's about
// to be appended.
func (q *Queue) OnBeforeAppend(f func(q *Queue, j *Job)) {
	q.cbBeforeAppend = append(q.cbBeforeAppend, f)
}
func (q *Queue) beforeAppend(j *Job) {
	for _, cb := range q.cbBeforeAppend {
		cb(q, j)
	}
}

// NewQueue creates an empty Queue.
func NewQueue() (q *Queue) {
	q = new(Queue)
	q.Jobs = make([]*Job, 0)
	return q
}
