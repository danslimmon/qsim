package qsim

import (
	"math/rand"
)

// A Job is the object that passes through a queueing system.
type Job struct {
	// JobId is a random nonnegative integer that can serve to uniquely
	// identify the job. Unless withId was set to false in NewJob(), in which
	// case JobId will be 0.
	JobId int64
	// Attrs contains user-defined, string-valued job attributes. One would use
	// this if the behavior of jobs in the system isn't uniform.
	Attrs map[string]string
}

// NewJob creates a new... wait for it... Job.
//
// If withId is set to true, the Job will have a random nonnegative integer
// assigned to JobId. The caller is expected to seed the PRNG.
func NewJob(withId bool) (j *Job) {
	j = new(Job)
	j.Attrs = make(map[string]string)
	if withId {
		j.JobId = rand.Int63()
	}
	return j
}
