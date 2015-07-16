package qsim

import (
	"math/rand"
)

// A Job is the object that passes through a queueing system.
type Job struct {
	// JobId is a random nonnegative integer that can serve to uniquely
	// identify the job.
	JobId int64
	// Attrs contains user-defined, string-valued job attributes. One would use
	// this if the behavior of jobs in the system isn't uniform.
	Attrs map[string]string
}

// NewJob creates a new... wait for it... Job.
//
// The Job will have a random nonnegative integer assigned to JobId. The caller
// is expected to seed the PRNG if necessary..
func NewJob() (j *Job) {
	j = new(Job)
	j.Attrs = make(map[string]string)
	j.JobId = rand.Int63()
	return j
}
