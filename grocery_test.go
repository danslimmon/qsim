package qsim

import (
	"fmt"
	"math/rand"
	"testing"
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

	SumQueued    int
	MaxQueued    int
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
		sys.processors[i] = NewProcessor()
		sys.processors[i].SetProcTimeGenerator(procTimeGenerator)
		sys.processors[i].AfterFinish(func(p *Processor, j *Job) {
			sys.FinishedJobs = append(sys.FinishedJobs, j)
		})
	}
	// When customers are ready to check out, they get in the shortest
	// queue. Unless there's an empty register, in which case they go
	// right ahead and start checking out.
	sys.arrBeh = NewShortestQueueArrBeh(sys.queues, sys.processors)
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
	// Add the current number of customers in the queue to sumQueued. We
	// are going to use this sum to generate the average at the end of
	// the simulation, so we need to weight it by the amount of time
	// elapsed since the last time we collected data.
	currentlyQueued := 0
	for _, q := range sys.queues {
		currentlyQueued += q.Length()
	}
	// Add the current number of customers in the queue to SumQueued. We
	// are going to use this sum to generate the average at the end of
	// the simulation, so we need to weight it by the amount of time
	// elapsed since the last time we collected data.
	sys.SumQueued += (clock - sys.prevClock) * currentlyQueued
	// Also keep track of the highest number of waiting customers we've
	// seen.
	if currentlyQueued > sys.MaxQueued {
		sys.MaxQueued = currentlyQueued
	}

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
// In this example we don't use it.
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
	var finalTick int

	// Run the simulation for a day (86400 seconds)
	sys := &GrocerySystem{}
	finalTick = RunSimulation(sys, 86400*1000)

	// Print our results out.
	fmt.Printf("Simulation lasted %0.3f seconds\n", float64(finalTick)/1000.0)
	fmt.Printf("Number of customers checked out: %d\n", sys.FinishedJobs)
	fmt.Printf("Average number of queued customers: %0.2f\n", float64(sys.SumQueued)/float64(finalTick))
	fmt.Printf("Highest number of queued customers: %d\n", sys.MaxQueued)
	fmt.Printf("Average time spent in system: %0.2f seconds\n", float64(sys.SumTotalTime)/float64(sys.NumFinishedJobs)/1000.0)
}
