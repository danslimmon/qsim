package qsim

import (
	"testing"
)

// Tests that JobId is set to a random nonnegative 63-bit integer.
func TestJobId(t *testing.T) {
	t.Parallel()
	var j *Job
	var jobs map[int64]*Job
	var i int
	var ok bool

	jobs = make(map[int64]*Job, 100)
	for i = 0; i < 100; i++ {
		j = NewJob()
		if j.JobId < 0 {
			t.Log("Got negative JobId", j.JobId)
			t.Fail()
		}
		if _, ok = jobs[j.JobId]; ok {
			t.Log("Got duplicate JobId", j.JobId)
			t.Fail()
		}
		jobs[j.JobId] = j
	}
}
