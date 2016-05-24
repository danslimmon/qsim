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
	D("Added event for time", newEv.T)

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

// Next returns the events in the schedule that are next to occur and removes
// those events from the schedule. It also returns the tick at which those
// events occur.
func (sch *Schedule) NextTick() (events []simEvent, tick int) {
	var i int

	// This should never happen, which means it definitely will some day.
	if len(sch.events) == 0 {
		panic("next Schedule event requested but Schedule is empty")
	}

	events = append(events, sch.events[0])
	for i = 1; i < len(sch.events); i++ {
		if sch.events[i].T == events[0].T {
			events = append(events, sch.events[i])
		} else {
			break
		}
	}
	sch.events = sch.events[i:]
	return events, events[0].T
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

// A System is the thing you simulate. You implement this interface and
// pass your implementation to RunSimulation().
type System interface {
	// Init runs before the simulation begins, and its job is to set up the
	// Queues, Processors, and behaviors.
	Init()
	// ArrProc returns the system's arrival process.
	ArrProc() ArrProc
	// ArrBeh returns the system's arrival behavior.
	ArrBeh() ArrBeh
	// BeforeTick is called right before the clock starts on a simulation.
	BeforeFirstTick()
	// BeforeEvents runs at every tick when a simulation event happens (a
	// Job arrives in the system, or a Job finishes processing and leaves
	// the system). BeforeEvents is called after all the events for the tick
	// in question have finished.
	BeforeEvents(clock int)
	// AfterEvents runs at every tick when a simulation event happens, but
	// in contrast with BeforeEvents, it runs after all the events for that
	// tick have occurred.
	AfterEvents(clock int)
	// Processors returns the list of Processors in the system.
	Processors() []*Processor
}

// RunSimulation simulates a queueing system for a certain number of ticks.
//
// The internal operations of a queuing system take care of themselves, so
// this function is only responsible for things going into and out of the
// system. It keeps track of the clock and triggers arrivals and
// job-finishes at the appropriate times.
//
// The return value is the last tick on which events occurred in the
// simulation. This may or may not be equal to maxTicks.
func RunSimulation(sys System, maxTicks int) (finalTick int) {
	var sch *Schedule
	var p *Processor
	var clock int
	var ev simEvent
	var events []simEvent

	sys.Init()
	sch = NewSchedule()

	// Schedule Processor-finish events. Each Processor gets an AfterStart
	// callback that schedules a Finish() call for that processor to occur
	// when the processing time has elapsed.
	cbAfterStart := func(cbProcessor *Processor, cbJob *Job, cbProcTime int) {
		eventCb := func(cbClock int) {
			cbProcessor.Finish()
		}
		sch.Add(simEvent{clock + cbProcTime, eventCb})
	}
	for _, p = range sys.Processors() {
		p.AfterStart(cbAfterStart)
	}

	// Schedule arrival events, including the initial one.
	cbAfterArrive := func(cbArrProc ArrProc, cbJobs []*Job, cbInterval int) {
		eventCb := func(cbClock int) {
			sys.ArrProc().Arrive(cbClock)
		}
		sch.Add(simEvent{clock + cbInterval, eventCb})
	}
	sys.ArrProc().AfterArrive(cbAfterArrive)
	sch.Add(simEvent{0, func(cbClock int) { sys.ArrProc().Arrive(cbClock) }})

	// Run the simulation.
	sys.BeforeFirstTick()
	for clock = 0; clock <= maxTicks; {
		events, clock = sch.NextTick()
		D()
		D("BEGIN TICK", clock)
		sys.BeforeEvents(clock)
		for _, ev = range events {
			ev.F(clock)
		}
		sys.AfterEvents(clock)
		D("END TICK", clock)
	}

	return clock
}
