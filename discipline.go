package qsim

type Discipline interface{}

// OneToOneFIFODiscipline moves Jobs from Queues to Processors based on
// the following algorithm:
//
// – There is a one-to-one relationship between Queues and Processors
// – When a Processor finishes a Job, the Job at the head of the
//   corresponding Queue is started on that Processor.
// – If a Processor finishes a Job when its corresponding Queue is
//   empty, it stays idle.
type OneToOneFIFODiscipline struct {
	Queues     []*Queue
	Processors []*Processor
}

// Generates a OneToOneFIFODiscipline given the Queues and Processors that
// should be linked to each other. queues and procs must be slices of equal
// length.
func NewOneToOneFIFODiscipline(queues []*Queue, procs []*Processor) (d Discipline) {
	var q *Queue
	var i int

	d = new(OneToOneFIFODiscipline)
	for i, q = range queues {
		cbAfterFinish := func(cbProc *Processor, cbJob *Job) {
			var j *Job
			j, _ = q.Shift()
			if j != nil {
				cbProc.Start(j)
			}
		}
		procs[i].AfterFinish(cbAfterFinish)
	}
	return
}
