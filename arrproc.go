package qsim

// An ArrProc (short for "arrival process") generates new Jobs at some interval.
type ArrProc interface {
	// Arrive generates a new job and also returns the interval that will elapse
	// before the next arrival.
	Arrive() (j *Job, interval int)
}

// ConstantArrProc generates jobs at a constant interval.
//
// It implements the ArrProc interface.
type ConstantArrProc struct {
	// Interval is the interval at which ConstantArrProc will generate Jobs.
	Interval int
}

// Generates a Job and returns the constant value of Interval.
func (ap *ConstantArrProc) Arrive() (j *Job, interval int) {
	j = NewJob()
	interval = ap.Interval
	return
}
