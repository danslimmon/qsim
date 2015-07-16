package qsim

import (
	"errors"
)

// A Processor is the piece of the queueing system that processes jobs.
type Processor struct {
	CurrentJob *Job

	procTimeGenerator func(j *Job) int
	// Callback lists
	cbBeforeStart  []func(q *Processor, j *Job)
	cbAfterStart   []func(q *Processor, j *Job, procTime int)
	cbBeforeFinish []func(q *Processor, j *Job)
	cbAfterFinish  []func(q *Processor, j *Job)
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
	p.beforeStart(j)
	if p.CurrentJob != nil {
		p.afterStart(nil, 0)
		return 0, errors.New("Tried to start job on busy processor; call Finish() first")
	}
	p.CurrentJob = j
	procTime = p.procTimeGenerator(j)
	p.afterStart(j, procTime)
	return procTime, nil
}

// Finish empties the current job out of the Processor and returns it.
//
// If Finish is called on an idle processor, j will be nil.
func (p *Processor) Finish() (j *Job) {
	j = p.CurrentJob
	p.beforeFinish(j)
	p.CurrentJob = nil
	return j
}

// BeforeStart adds a callback to be run immediately before a Job is started
// on the processor.
//
// The callback will be passed the processor itself and the job that's about
// to be started. If Start is called on a busy Processor, there is no change
// in the callback's behavior: it's still passed the new job, but the job won't
// actually get started.
func (p *Processor) BeforeStart(f func(p *Processor, j *Job)) {
	p.cbBeforeStart = append(p.cbBeforeStart, f)
}
func (p *Processor) beforeStart(j *Job) {
	for _, cb := range p.cbBeforeStart {
		cb(p, j)
	}
}

// AfterStart adds a callback to be run immediately after a Job is
// started on the processor.
//
// The callback will be passed the processor itself, the job that was
// just started, and the processing time that was decided upon for the
// job. If Start is called on an busy processor, this callback will
// run but j will be nil.
func (p *Processor) AfterStart(f func(p *Processor, j *Job, procTime int)) {
	p.cbAfterStart = append(p.cbAfterStart, f)
}
func (p *Processor) afterStart(j *Job, procTime int) {
	for _, cb := range p.cbAfterStart {
		cb(p, j, procTime)
	}
}

// BeforeFinish adds a callback to be run immediately before a Job is finished
// on the processor.
//
// The callback will be passed the processor itself and the job that's about
// to be finished. If Finish is called on an idle Processor, the callback still
// runs but j is nil.
func (p *Processor) BeforeFinish(f func(p *Processor, j *Job)) {
	p.cbBeforeFinish = append(p.cbBeforeFinish, f)
}
func (p *Processor) beforeFinish(j *Job) {
	for _, cb := range p.cbBeforeFinish {
		cb(p, j)
	}
}

// NewProcessor creates a new Processor struct.
func NewProcessor() (p *Processor) {
	return new(Processor)
}
