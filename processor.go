package qsim

import (
	"errors"
)

// A Processor is the piece of the queueing system that processes jobs.
type Processor struct {
	CurrentJob        *Job
	procTimeGenerator func(j *Job) int
}

// SetProcTimeGenerator sets the function that will generate processing
// times for jobs.
//
// For example, if you wanted half of the jobs to take 10 ticks to
// process, and the other half to take 20 ticks, you could do this:
//
//    ptg := func(j *Job) int {
//        if rand.Float32() < 0.5 {
//            return 10
//        } else {
//            return 20
//        }
//    }
//    p.SetProcTimeGenerator(ptg)
func (p *Processor) SetProcTimeGenerator(ptg func(j *Job) int) {
	p.procTimeGenerator = ptg
}

// Start begins processing a given job.
//
// The return value is the amount of time it'll take to process the job.
// This method will throw an error if called when there's already a job
// being processed: that job needs to be Finish()ed first.
func (p *Processor) Start(j *Job) (procTime int, err error) {
	if p.CurrentJob != nil {
		return 0, errors.New("Tried to start job on busy processor; call Finish() first")
	}
	p.CurrentJob = j
	return p.procTimeGenerator(j), nil
}

// Finish empties the current job out of the Processor and returns it.
//
// If Finish is called on an idle processor, j will be nil.
func (p *Processor) Finish() (j *Job) {
	j = p.CurrentJob
	p.CurrentJob = nil
	return j
}

// NewProcessor creates a new Processor struct.
func NewProcessor() (p *Processor) {
	return new(Processor)
}
