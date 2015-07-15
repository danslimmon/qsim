package qsim

// A Job is the object that passes through a queueing system.
type Job struct {
	// Attrs contains user-defined, string-valued job attributes.
	//
	// One would use this if the behavior of jobs in the system isn't uniform.
	Attrs map[string]string
}

// NewJob creates a new... wait for it... Job.
func NewJob() (j *Job) {
	j = new(Job)
	j.Attrs = make(map[string]string)
	return j
}
