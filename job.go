package qsim

import (
	"math/rand"
)

// A Job is the object that passes through a queueing system.
type Job struct {
	// JobId is a random nonnegative integer that can serve to uniquely
	// identify the job.
	JobId int64
	// The time the Job arrived in the system.
	ArrTime int
	// Attrs contains user-defined, string-valued job attributes. One would use
	// this if the behavior of jobs in the system isn't uniform.
	Attrs map[string]string
}

// NewJob creates a new... wait for it... Job.
//
// arrTime should be the simulation clock time at which the Job arrived.
//
// The Job will have a random nonnegative integer assigned to JobId. The caller
// is expected to seed the PRNG if necessary..
func NewJob(arrTime int) (j *Job) {
	j = new(Job)
	j.Attrs = make(map[string]string)
	j.JobId = rand.Int63()
	j.ArrTime = arrTime
	return j
}
