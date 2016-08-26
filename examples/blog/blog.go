package main

/* I created this system to generate diagrams for a blog post about the relationship
 * between utilization and queue size:
 *
 * https://danslimmon.wordpress.com/2016/08/26/the-most-important-thing-to-understand-about-queues/
 */

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/danslimmon/qsim"
)

// To run a simulation, you have to implement the System interface:
// https://godoc.org/github.com/danslimmon/qsim#System
type BlogSystem struct {
	// The list of all queues in the system.
	queues []*qsim.Queue
	// The list of all processors in the system.
	processors []*qsim.Processor
	// The system's arrival process
	arrProc qsim.ArrProc
	// The system's arrival behavior
	arrBeh qsim.ArrBeh

	IdleTime        int
	ArrivalInterval float64
	QueueSum        int
	QueueCount      int

	prevClock int
}

// Init runs before the simulation begins, and its job is to set up the
// queues, processors, and behaviors.
func (sys *BlogSystem) Init() {
	var i int
	rand.Seed(time.Now().UnixNano())
	sys.arrProc = qsim.NewPoissonArrProc(sys.ArrivalInterval)
	procTimeGenerator := func(j *qsim.Job) int {
		return int(rand.ExpFloat64() * 1000.0)
	}
	// There is 1 processor and 1 queue
	sys.queues = make([]*qsim.Queue, 1)
	sys.processors = make([]*qsim.Processor, 1)
	sys.queues[i] = qsim.NewQueue()
	sys.processors[i] = qsim.NewProcessor(procTimeGenerator)
	sys.arrBeh = qsim.NewShortestQueueArrBeh(sys.queues, sys.processors, sys.arrProc)
	qsim.NewOneToOneFIFODiscipline(sys.queues, sys.processors)
}

// ArrProc returns the system's arrival process.
func (sys *BlogSystem) ArrProc() qsim.ArrProc {
	return sys.arrProc
}

// ArrBeh returns the system's arrival behavior.
func (sys *BlogSystem) ArrBeh() qsim.ArrBeh {
	return sys.arrBeh
}

func (sys *BlogSystem) BeforeFirstTick() {}

// Processors returns the list of Processors in the system.
func (sys *BlogSystem) Processors() []*qsim.Processor {
	return sys.processors
}

func (sys *BlogSystem) BeforeEvents(clock int) {
	sys.QueueSum += sys.queues[0].Length()
	sys.QueueCount++
	if sys.processors[0].IsIdle() {
		sys.IdleTime += clock - sys.prevClock
	}
	sys.prevClock = clock
}

func (sys *BlogSystem) AfterEvents(clock int) {}

func main() {
	var finalTick, simTicks int

	// Run the simulation for 24 hours (a tick represents a millisecond)
	simTicks = 86400 * 1000

	fmt.Printf("arrival_interval,utilization,avg_queue\n")
	for ai := 900; ai < 3000; ai += 50 {
		sys := &BlogSystem{ArrivalInterval: float64(ai)}
		finalTick = qsim.RunSimulation(sys, simTicks)
		fmt.Printf("%d,%0.3f,%0.3f\n",
			ai, 1.0-float64(sys.IdleTime)/float64(finalTick), float64(sys.QueueSum)/float64(sys.QueueCount))
	}
}
