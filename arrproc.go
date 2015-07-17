package qsim

// An ArrProc (short for "arrival process") generates new Jobs at some interval.
type ArrProc interface {
	Arrive() (j *Job, interval int)
	BeforeArrive(f func(ap ArrProc))
}

// ConstantArrProc generates jobs at a constant interval.
//
// It implements the ArrProc interface.
type ConstantArrProc struct {
	// Interval is the interval at which ConstantArrProc will generate Jobs.
	Interval int

	// Callback lists
	cbBeforeArrive []func(ap ArrProc)
}

// Generates a Job and returns the constant value of Interval as the number of
// ticks that will elapse before the next arrival.
func (ap *ConstantArrProc) Arrive() (j *Job, interval int) {
	ap.beforeArrive()
	j = NewJob()
	interval = ap.Interval
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

// NewConstantArrProc returns a new ConstantArrProc with the given Interval
// value.
func NewConstantArrProc(interval int) (ap *ConstantArrProc) {
	ap = new(ConstantArrProc)
	ap.Interval = interval
	return ap
}
