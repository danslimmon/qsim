package qsim

import (
	"testing"
)

func TestSchedule(t *testing.T) {
	t.Parallel()
	var sch *Schedule
	var ev simEvent
	var events []simEvent
	var addOrder []int
	var tick int
	type recvExpectation struct{ Tick, NumEvents int }
	var recvOrder []recvExpectation
	var exp recvExpectation
	f := func(clock int) {}

	sch = NewSchedule()
	addOrder = []int{3, 5, 5, 2, 10, 8}
	for _, tick = range addOrder {
		sch.Add(simEvent{tick, f})
	}

	recvOrder = []recvExpectation{
		{2, 1},
		{3, 1},
		{5, 2},
		{8, 1},
		{10, 1},
	}
	for _, exp = range recvOrder {
		events, tick = sch.NextTick()
		if len(events) != exp.NumEvents {
			t.Log("Expected", exp.NumEvents, "events at tick", tick, "but got", len(events))
			t.Fail()
		}
		for _, ev = range events {
			if ev.T != tick {
				t.Log("Expected the next scheduled event to have T =", tick, "but got", ev.T)
				t.Fail()
			}
		}
	}
}
