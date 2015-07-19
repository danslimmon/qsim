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
	// IntAttrs contains user-defined, int-valued job attributes. You can use
	// this feature for testing, or for debugging, or for changing the behavior
	// of the system for particular types of jobs.
	IntAttrs map[string]int
	// StrAttrs contains user-defined, str-valued job attributes. You can use
	// this feature for testing, or for debugging, or for changing the behavior
	// of the system for particular types of jobs.
	StrAttrs map[string]string
}

// NewJob creates a new... wait for it... Job.
//
// arrTime should be the simulation clock time at which the Job arrived.
//
// The Job will have a random nonnegative integer assigned to JobId. The caller
// is expected to seed the PRNG if necessary..
func NewJob(arrTime int) (j *Job) {
	j = new(Job)
	j.IntAttrs = make(map[string]int)
	j.StrAttrs = make(map[string]string)
	j.JobId = rand.Int63()
	j.ArrTime = arrTime
	return j
}
