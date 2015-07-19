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

func (d *OneToOneFIFODiscipline) AssignQueueToProcessor(q *Queue, p *Processor) {
	cbAfterFinish := func(cbProc *Processor, cbJob *Job) {
		var j *Job
		j, _ = q.Shift()

		if j != nil {
			cbProc.Start(j)
		}

		// Debug output
		if j == nil {
			D("Processor", cbProc.ProcessorId, "finished job", cbJob.JobId, "and now its Queue", q.QueueId, "is empty")
		} else {
			D("Processor", cbProc.ProcessorId, "finished job", cbJob.JobId, "and began Job", j.JobId, "from Queue", q.QueueId)
		}
	}
	p.AfterFinish(cbAfterFinish)
}

// Generates a OneToOneFIFODiscipline given the Queues and Processors that
// should be linked to each other. queues and procs must be slices of equal
// length.
func NewOneToOneFIFODiscipline(queues []*Queue, procs []*Processor) Discipline {
	var i int
	var d *OneToOneFIFODiscipline

	d = new(OneToOneFIFODiscipline)
	for i, _ = range queues {
		d.AssignQueueToProcessor(queues[i], procs[i])
	}
	return d
}
