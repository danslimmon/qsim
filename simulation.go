package qsim

// An event scheduled to occur in the simulation We'll run the function F at
// tick Time. F will be called at time Time with the current clock time as its
// argument.
type simEvent struct {
	T int
	F func(clock int)
}

// Schedule holds simEvents in the order that they need to be run. Events that
// have already occurred are removed.
type Schedule struct {
	// The list of events that have yet to occur. This is kept in ascending time
	// order.
	events []simEvent
}

// Add puts a new event in the schedule.
func (sch *Schedule) Add(newEv simEvent) {
	var i int

	if len(sch.events) == 0 {
		sch.events = append(sch.events, newEv)
		return
	}

	for i = len(sch.events) - 1; i >= 0; i-- {
		if sch.events[i].T <= newEv.T {
			sch.insertEvent(i+1, newEv)
			return
		}
	}
	// Fell off the beginning of the schedule, so just insert at the beginning
	sch.insertEvent(0, newEv)
	return
}

// Next returns the next event in the schedule and removes that event from the
// schedule.
func (sch *Schedule) Next() simEvent {
	var ev simEvent

	// This should never happen, which means it definitely will.
	if len(sch.events) == 0 {
		panic("next Schedule event requested but Schedule is empty")
	}

	ev = sch.events[0]
	sch.events = sch.events[1:]
	return ev
}

// insertEvent places a simEvent at the given index in sch.events.
func (sch *Schedule) insertEvent(idx int, newEv simEvent) {
	sch.events = append(sch.events, simEvent{})
	copy(sch.events[idx+1:], sch.events[idx:])
	sch.events[idx] = newEv
	return
}

// NewSchedule creates a new, empty Schedule struct.
func NewSchedule() *Schedule {
	return new(Schedule)
}

// RunSimulation simulates a queueing system for a certain number of ticks.
//
// The internal operations of a queuing system take care of themselves, so
// this function is only responsible for things going in and out of the
// system. It keeps track of the clock and triggers arrivals and
// job-finishes at the appropriate times.
/*func RunSimulation(ap ArrProc, procs []*Processor, ticks int) {
	var schedule *Schedule
	var clock int
}*/
