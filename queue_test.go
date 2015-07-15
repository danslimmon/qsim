package qsim

import (
	"fmt"
	"testing"
)

// Tests serial queue insertion
func TestQueueAppend(t *testing.T) {
	t.Parallel()
	var q *Queue
	var j *Job
	q = NewQueue()

	for i := 0; i < 20; i++ {
		j = NewJob()
		j.Attrs["i"] = fmt.Sprintf("%d", i)
		q.Append(j)
	}

	if q.Jobs[0].Attrs["i"] != "0" {
		t.Log("Zeroeth element not found at index 0")
		t.Fail()
	}
	if q.Jobs[19].Attrs["i"] != "19" {
		t.Log("Nineteenth element not found at index 19")
		t.Fail()
	}
	if len(q.Jobs) != 20 {
		t.Fail()
	}
}
