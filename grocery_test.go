package qsim

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

// To run a simulation, you have to implement the System interface:
// https://godoc.org/github.com/danslimmon/qsim#System
type GrocerySystem struct {
	// The list of all queues in the system.
	queues []*Queue
	// The list of all processors in the system.
	processors []*Processor
	// The system's arrival process
	arrProc ArrProc
	// The system's arrival behavior
	arrBeh ArrBeh

	SumCustomers int
	SumTotalTime int
	// Holds the list of Jobs that have finished since the last tick. We
	// use this to keep track of the total time spent by customers in the
	// system.
	FinishedJobs    []*Job
	NumFinishedJobs int
	prevClock       int
}

// Init runs before the simulation begins, and its job is to set up the
// queues, processors, and behaviors.
func (sys *GrocerySystem) Init() {
	var i int
	rand.Seed(time.Now().UnixNano())
	// Customers arrive at the checkout line an average of every 30 seconds
	// and the intervals between their arrivals are exponentially
	// distributed.
	sys.arrProc = NewPoissonArrProc(30000.0)
	// The time taken to check a customer out is normally distributed, with
	// a mean of 60 seconds and a standard deviation of 10 seconds.
	procTimeGenerator := func(j *Job) int {
		return int(rand.NormFloat64()*10000.0 + 60000.0)
	}
	// There are 3 registers and 3 queues.
	sys.queues = make([]*Queue, 3)
	sys.processors = make([]*Processor, 3)
	for i = 0; i < 3; i++ {
		sys.queues[i] = NewQueue()
		sys.queues[i].QueueId = i
		sys.processors[i] = NewProcessor(procTimeGenerator)
		sys.processors[i].ProcessorId = i
		sys.processors[i].AfterFinish(func(p *Processor, j *Job) {
			sys.FinishedJobs = append(sys.FinishedJobs, j)
		})
	}
	// When customers are ready to check out, they get in the shortest
	// queue. Unless there's an empty register, in which case they go
	// right ahead and start checking out.
	sys.arrBeh = NewShortestQueueArrBeh(sys.queues, sys.processors, sys.arrProc)
	// Customers stay in the queue they originally joined, and each queue
	// leads to exactly one register.
	NewOneToOneFIFODiscipline(sys.queues, sys.processors)
}

// ArrProc returns the system's arrival process.
func (sys *GrocerySystem) ArrProc() ArrProc {
	return sys.arrProc
}

// ArrBeh returns the system's arrival behavior.
func (sys *GrocerySystem) ArrBeh() ArrBeh {
	return sys.arrBeh
}

func (sys *GrocerySystem) BeforeFirstTick() {}

// BeforeEvents runs at every tick when a simulation event happens (a
// Job arrives in the system, or a Job finishes processing and leaves
// the system). BeforeEvents is called after all the events for the tick
// in question have finished.
//
// In this example, we use BeforeEvents to calculate stats about the
// system.
func (sys *GrocerySystem) BeforeEvents(clock int) {
	// Ignore the initial tick.
	if clock == 0 {
		return
	}
	// Add the current number of customers in the system to
	// currentCustomers. We are going to use this sum to generate the
	// average at the end of the simulation, so we need to weight it
	// by the amount of time elapsed since the last time we collected
	// data.
	currentCustomers := 0
	currentlyQueued := 0
	for _, q := range sys.queues {
		currentCustomers += q.Length()
		currentlyQueued += q.Length()
	}
	for _, p := range sys.processors {
		if !p.IsIdle() {
			currentCustomers++
		}
	}
	// Add the current number of customers in the queue to SumCustomers.
	// We are going to use this sum to generate the average at the end
	// of the simulation, so we need to weight it by the amount of time
	// elapsed since the last time we collected data.
	sys.SumCustomers += (clock - sys.prevClock) * currentCustomers

	sys.prevClock = clock
}

// Processors returns the list of Processors in the system.
func (sys *GrocerySystem) Processors() []*Processor {
	return sys.processors
}

// AfterEvents runs at every tick when a simulation event happens, but
// in contrast with BeforeEvents, it runs after all the events for that
// tick have occurred.
//
// In this example we used it to keep track of the average time Jobs
// spend in the system (by calculating total Job-ticks and the number
// of Jobs finished).
func (sys *GrocerySystem) AfterEvents(clock int) {
	var j *Job
	if len(sys.FinishedJobs) != 0 {
		for _, j = range sys.FinishedJobs {
			sys.SumTotalTime += clock - j.ArrTime
			sys.NumFinishedJobs++
		}
		sys.FinishedJobs = sys.FinishedJobs[:0]
	}
}

// Simulates a small grocery store checkout line:
//
// – Customers arrive at the checkout line by a Poisson process (i.e. the
//   distribution of the time between arrivals is exponential). This is
//   probably a pretty good guess for real-world grocery stores.
// - There are 3 registers, each with its own queue. Once a customer enters
//   a queue, they stay in it until that register is empty.
// - The time taken to check a customer out is drawn from a normal
//   distribution.
// - Each tick is a millisecond (we use very small ticks to minimize the
//   rounding error inherent in picking integer times from a continuous
//   distribution.
func TestGrocery(t *testing.T) {
	var finalTick, simTicks int
	var avgOccupancy, avgArrivalRate, avgWait, precision float64

	// Run the simulation for a week
	simTicks = 7 * 86400 * 1000
	// Satisfy Little's Law to within 1 part in 1000
	precision = .001

	sys := &GrocerySystem{}
	finalTick = RunSimulation(sys, simTicks)

	// Make sure the simulation ran as long as it should have
	if finalTick < simTicks {
		t.Log("Simulation was supposed to run for", simTicks, "ticks but only ran for", finalTick)
		t.Fail()
	}

	// Make sure Little's Law holds.
	avgOccupancy = float64(sys.SumCustomers) / float64(finalTick)
	avgWait = float64(sys.SumTotalTime) / float64(sys.NumFinishedJobs)
	avgArrivalRate = float64(sys.NumFinishedJobs) / float64(finalTick)
	if math.Abs(avgArrivalRate*avgWait-avgOccupancy) > precision*avgOccupancy {
		t.Log("Little's law doesn't hold for GrocerySystem: average occupancy should be near", avgArrivalRate*avgWait, "but it is", avgOccupancy)
		t.Fail()
	}
}
