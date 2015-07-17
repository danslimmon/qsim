package qsim

import (
	"testing"
)

func TestSchedule(t *testing.T) {
	t.Parallel()
	var sch *Schedule
	var ev simEvent
	var addOrder, recvOrder []int
	var tick int
	f := func(clock int) {}

	sch = NewSchedule()
	addOrder = []int{3, 5, 5, 2, 10, 8}
	for _, tick = range addOrder {
		sch.Add(simEvent{tick, f})
	}

	recvOrder = []int{2, 3, 5, 5, 8, 10}
	for _, tick = range recvOrder {
		ev = sch.Next()
		if ev.T != tick {
			t.Log("Expected the next scheduled event to have T =", tick, "but got", ev.T)
			t.Fail()
		}
	}
}
