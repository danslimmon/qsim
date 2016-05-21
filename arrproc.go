package qsim

import (
	"math/rand"
)

// An ArrProc (short for "arrival process") generates new Jobs at some interval.
type ArrProc interface {
	Arrive(clock int) (jobs []*Job, interval int)
	BeforeArrive(f func(ap ArrProc))
	AfterArrive(f func(ap ArrProc, jobs []*Job, interval int))
}

// ConstantArrProc generates jobs at a constant interval.
//
// It implements the ArrProc interface.
type ConstantArrProc struct {
	// Interval is the interval at which ConstantArrProc will generate Jobs.
	Interval int

	// Callback lists
	cbBeforeArrive []func(ap ArrProc)
	cbAfterArrive  []func(ap ArrProc, jobs []*Job, interval int)
}

// Arrive generates a Job and returns the constant value of Interval as the
// number of ticks that will elapse before the next arrival.
//
// clock is the current simulation clock time.
func (ap *ConstantArrProc) Arrive(clock int) (jobs []*Job, interval int) {
	ap.beforeArrive()
	jobs = append(jobs, NewJob(clock))
	interval = ap.Interval
	ap.afterArrive(jobs, interval)
	return
}

// BeforeArrive adds a callback to run immediately before the Arrival Process
// creates a job. This callback is passed the ArrProc itself.
func (ab *ConstantArrProc) BeforeArrive(f func(ArrProc)) {
	ab.cbBeforeArrive = append(ab.cbBeforeArrive, f)
}
func (ab *ConstantArrProc) beforeArrive() {
	for _, cb := range ab.cbBeforeArrive {
		cb(ab)
	}
}

// AfterArrive adds a callback to run immediately after the Arrival Process
// creates a job. This callback is passed the ArrProc itself, the Jobs that
// were created, and the interval that will elapse before the next arrival.
func (ab *ConstantArrProc) AfterArrive(f func(ArrProc, []*Job, int)) {
	ab.cbAfterArrive = append(ab.cbAfterArrive, f)
}
func (ab *ConstantArrProc) afterArrive(jobs []*Job, interval int) {
	for _, cb := range ab.cbAfterArrive {
		cb(ab, jobs, interval)
	}
}

// NewConstantArrProc returns a new ConstantArrProc with the given Interval
// value.
func NewConstantArrProc(interval int) (ap *ConstantArrProc) {
	ap = new(ConstantArrProc)
	ap.Interval = interval
	return ap
}

// PoissonArrProc generates jobs according to a Poisson process. This means
// that the distribution of intervals between arrivals is an exponential
// distribution. (If you're wondering, lambda = 1. If not, don't worry about
// it.)
//
// Poisson processes useful for modeling radioactive decay, telephone calls at
// a call center, document requests on a web server, and many other punctual
// phenomena where events occur independently from each other.
//
// The distribution of arrival intervals is a continuous distribution, but we
// quantize time, so to minimize approximation errors you should make sure
// that your tick length is much smaller than your average arrival time. If
// jobs are normally arriving once every few seconds, then your ticks should
// be milliseconds.
//
// PoissonArrProc implements the ArrProc interface.
type PoissonArrProc struct {
	Mean float64

	// Callback lists
	cbBeforeArrive []func(ap ArrProc)
	cbAfterArrive  []func(ap ArrProc, jobs []*Job, interval int)
}

// Arrive generates a Job and returns the interval that will elapse before the
// subsequent arrival. These arrival intervals are exponentially distributed.
//
// clock is the current simulation clock time.
func (ap *PoissonArrProc) Arrive(clock int) (jobs []*Job, interval int) {
	ap.beforeArrive()
	jobs = append(jobs, NewJob(clock))
	interval = ap.pickInterval()
	ap.afterArrive(jobs, interval)
	return
}

// BeforeArrive adds a callback to run immediately before the Arrival Process
// creates a job. This callback is passed the ArrProc itself.
func (ap *PoissonArrProc) BeforeArrive(f func(ArrProc)) {
	ap.cbBeforeArrive = append(ap.cbBeforeArrive, f)
}
func (ap *PoissonArrProc) beforeArrive() {
	for _, cb := range ap.cbBeforeArrive {
		cb(ap)
	}
}

// AfterArrive adds a callback to run immediately after the Arrival Process
// creates a job. This callback is passed the ArrProc itself, the Jobs that
// were created, and the interval that will elapse before the next arrival.
func (ap *PoissonArrProc) AfterArrive(f func(ArrProc, []*Job, int)) {
	ap.cbAfterArrive = append(ap.cbAfterArrive, f)
}
func (ap *PoissonArrProc) afterArrive(jobs []*Job, interval int) {
	for _, cb := range ap.cbAfterArrive {
		cb(ap, jobs, interval)
	}
}

// Picks an arrival interval from an exponential distribution.
func (ab *PoissonArrProc) pickInterval() int {
	var r float64
	r = rand.ExpFloat64() * ab.Mean
	return int(r)
}

func NewPoissonArrProc(mean float64) (ap *PoissonArrProc) {
	return &PoissonArrProc{Mean: mean}
}
