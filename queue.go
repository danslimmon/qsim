package qsim

// A Queue holds Jobs until they're ready for processing.
type Queue struct {
	Jobs []*Job
}

// Append adds a Job to the end of the queue.
func (q *Queue) Append(j *Job) {
	q.Jobs = append(q.Jobs, j)
}

// NewQueue creates an empty Queue.
func NewQueue() (q *Queue) {
	q = new(Queue)
	q.Jobs = make([]*Job, 0)
	return q
}
