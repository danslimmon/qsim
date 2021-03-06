package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/danslimmon/qsim"
)

type PortaPottySystem struct {
	// The probability of a given person using the strategy
	PStrategy float64
	// When to start capturing stats. We use this to avoid sampling the
	// initial ramp-up of the system.
	StatsStart int

	// The list of all queues in the system.
	queues []*qsim.Queue
	// The list of all processors in the system.
	processors []*qsim.Processor
	// The system's arrival process
	arrProc qsim.ArrProc
	// The system's arrival behavior
	arrBeh qsim.ArrBeh

	SumStrategizerWaits, SumNonStrategizerWaits int
	NumStrategizers, NumNonStrategizers         int

	statsStarted bool
	finishedJobs []*qsim.Job
	prevClock    int
}

// Init runs before the simulation begins, and its job is to set up the
// queues, processors, and behaviors.
func (sys *PortaPottySystem) Init() {
	var i int
	var maleMean, femaleMean, stdev float64

	rand.Seed(time.Now().UnixNano())
	maleMean = 40000.0
	femaleMean = 60000.0
	stdev = 5000.0

	// The time taken to use the porta-potty depends on the sex of the
	// person using it.
	procTimeGenerator := func(j *qsim.Job) int {
		if j.StrAttrs["sex"] == "male" {
			// Normal distribution of pee times with stdev=5s
			return int(rand.NormFloat64()*stdev + maleMean)
		} else {
			return int(rand.NormFloat64()*stdev + femaleMean)
		}
	}

	// There are 15 porta-potties, each with its own queue.
	sys.queues = make([]*qsim.Queue, 15)
	sys.processors = make([]*qsim.Processor, 15)
	for i = 0; i < 15; i++ {
		sys.queues[i] = qsim.NewQueue()
		sys.queues[i].MaxLength = 8
		sys.processors[i] = qsim.NewProcessor(procTimeGenerator)
	}

	// Processor callback to keep track of wait times for strategy users and non
	// strategy users. This saves finished Jobs to a slice so that we can do
	// calculations on those Jobs after each tick in AfterEvents.
	for i, _ = range sys.processors {
		sys.processors[i].AfterFinish(func(p *qsim.Processor, j *qsim.Job) {
			if !sys.statsStarted {
				return
			}
			sys.finishedJobs = append(sys.finishedJobs, j)
		})
	}

	// The mean of this Poisson distribution is the maxmimum rate at which
	// porta-potties can be vacated. This ensures that queues will usually be
	// long.
	sys.arrProc = qsim.NewPoissonArrProc((maleMean + femaleMean) / 2.0 / 15.0)
	// Assign a gender to each incoming person.
	sys.arrProc.AfterArrive(func(ap qsim.ArrProc, jobs []*qsim.Job, interval int) {
		sexes := []string{"male", "female"}
		jobs[0].StrAttrs["sex"] = sexes[rand.Intn(2)]
	})
	// Occasionally pick a person to use the strategy.
	sys.arrProc.AfterArrive(func(ap qsim.ArrProc, jobs []*qsim.Job, interval int) {
		if rand.Float64() < sys.PStrategy {
			jobs[0].IntAttrs["use_strategy"] = 1
		} else {
			jobs[0].IntAttrs["use_strategy"] = 0
		}
	})

	// When customers arrive, they pick the shortest queue. If all the queues
	// are too long, the queue MaxLength parameter means they go do something
	// else.
	sys.arrBeh = qsim.NewShortestQueueArrBeh(sys.queues, sys.processors, sys.arrProc)
	// This callback overrides the default arrival behavior for people that
	// are using our clever strategy.
	sys.arrBeh.BeforeAssign(func(ab qsim.ArrBeh, j *qsim.Job) *qsim.Assignment {
		if j.IntAttrs["use_strategy"] == 1 {
			return sys.strategicAssignment()
		} else {
			return nil
		}
	})

	// Customers stay in the queue they originally joined, and each queue
	// leads to exactly one porta-potty.
	qsim.NewOneToOneFIFODiscipline(sys.queues, sys.processors)
}

// ArrProc returns the system's arrival process.
func (sys *PortaPottySystem) ArrProc() qsim.ArrProc {
	return sys.arrProc
}

// ArrBeh returns the system's arrival behavior.
func (sys *PortaPottySystem) ArrBeh() qsim.ArrBeh {
	return sys.arrBeh
}

// Occupancy returns the total number of Jobs in the system.
func (sys *PortaPottySystem) Occupancy() (occ int) {
	var p *qsim.Processor
	var q *qsim.Queue
	for _, p = range sys.processors {
		if !p.IsIdle() {
			occ++
		}
	}
	for _, q = range sys.queues {
		occ += q.Length()
	}
	return
}

// Processors returns the list of Processors in the system.
func (sys *PortaPottySystem) Processors() []*qsim.Processor {
	return sys.processors
}

func (sys *PortaPottySystem) BeforeFirstTick() {}

// BeforeEvents runs at every tick when a simulation event happens (a
// Job arrives in the system, or a Job finishes processing and leaves
// the system). BeforeEvents is called after all the events for the tick
// in question have finished.
func (sys *PortaPottySystem) BeforeEvents(clock int) {}

// AfterEvents runs at every tick when a simulation event happens, but
// in contrast with BeforeEvents, it runs after all the events for that
// tick have occurred.
func (sys *PortaPottySystem) AfterEvents(clock int) {
	var j *qsim.Job

	// Ignore the initial transient behavior of the system
	if clock < sys.StatsStart {
		sys.prevClock = clock
		return
	}
	sys.statsStarted = true

	for _, j = range sys.finishedJobs {
		if j.IntAttrs["use_strategy"] == 1 {
			sys.SumStrategizerWaits += clock - j.ArrTime
			sys.NumStrategizers++
		} else {
			sys.SumNonStrategizerWaits += clock - j.ArrTime
			sys.NumNonStrategizers++
		}
	}
	sys.finishedJobs = sys.finishedJobs[:0]
	sys.prevClock = clock
}

// strategicAssignment returns an Assignment corresponding to the following
// wait-time reduction strategy:
//
// – Look for the shortest queue first. If there are multiple queues of the
//   same short length, then
// – Pick the queue with the highest male-to-female ratio. If there are
//   multiple short queues with the same male-to-female ratio, then
// – Pick a random one.
func (sys *PortaPottySystem) strategicAssignment() *qsim.Assignment {
	var p *qsim.Processor
	var q *qsim.Queue
	var shortQueues, dudefulQueues []*qsim.Queue
	var j *qsim.Job
	var shortestLen, maleCount, highestMaleCount int
	shortestLen = 9999

	// If there's an idle porta-potty, obviously just walk in
	for _, p = range sys.processors {
		if p.IsIdle() {
			return nil
		}
	}

	for _, q = range sys.queues {
		if q.Length() < shortestLen {
			shortQueues = []*qsim.Queue{q}
			shortestLen = q.Length()
		} else if q.Length() == shortestLen {
			shortQueues = append(shortQueues, q)
		}
	}

	for _, q = range shortQueues {
		maleCount = 0
		for _, j = range q.Jobs {
			if j.StrAttrs["sex"] == "male" {
				maleCount++
			}
		}
		if maleCount > highestMaleCount {
			dudefulQueues = []*qsim.Queue{q}
			highestMaleCount = maleCount
		} else if maleCount == highestMaleCount {
			dudefulQueues = append(dudefulQueues, q)
		}
	}

	return &qsim.Assignment{
		Type:  "Queue",
		Queue: dudefulQueues[rand.Intn(len(dudefulQueues))],
	}
}

// Simulates a line of porta-potties at a big concert.
//
// – People arrive very frequently, but if the queues are too long they leave.
// – There are 15 porta-potties, each with its own queue. Once a person enters
//   a queue, they stay in it until that porta-potty is vacant.
// – The time taken to use a porta-potty is normally distributed with a
//   different mean for men and women.
// – Most people just pick a random queue to join (as long as it's no longer
//   than the shortest queue), but some people use the strategy of getting
//   into the queue with the highestman:woman ratio (again, as long as it's no
//   longer than the shortest queue), on the theory that this will get them to
//   the front of the queue faster.
// – Each tick is a millisecond (we use very small ticks to minimize the
//   rounding error inherent in picking integer times from a continuous
//   distribution.
func SimPortaPotty() {
	var simTicks, simsPerProb int
	var probStep float64
	type simResult struct {
		Done                                        bool
		PStrategy                                   float64
		SumStrategizerWaits, SumNonStrategizerWaits int
		NumStrategizers, NumNonStrategizers         int
	}
	var ch chan simResult
	var cpu, nCpu, nProbs, probsPerCpu, routinesDone int

	fmt.Println("pStrategy,avgStratWait,avgNonStratWait,avgWait")

	nCpu = 5
	nProbs = 100
	probStep = .01
	probsPerCpu = nProbs / nCpu
	// Run each simulation for 14 days
	simTicks = 14 * 86400 * 1000
	simsPerProb = 40

	ch = make(chan simResult)
	for cpu = 0; cpu < nCpu; cpu++ {
		go func(cpu int) {
			var i int
			for i = cpu*probsPerCpu + 1; i <= (cpu+1)*probsPerCpu; i++ {
				var pStrategy float64
				var rslt simResult
				pStrategy = probStep * float64(i)
				rslt = simResult{
					Done:      false,
					PStrategy: pStrategy,
				}
				var j int
				for j = 0; j < simsPerProb; j++ {
					sys := &PortaPottySystem{
						PStrategy:  pStrategy,
						StatsStart: 200000000,
					}
					qsim.RunSimulation(sys, simTicks)

					rslt.SumStrategizerWaits += sys.SumStrategizerWaits
					rslt.SumNonStrategizerWaits += sys.SumNonStrategizerWaits
					rslt.NumStrategizers += sys.NumStrategizers
					rslt.NumNonStrategizers += sys.NumNonStrategizers
				}
				ch <- rslt
			}
			ch <- simResult{Done: true}
		}(cpu)
	}

	for routinesDone < nCpu {
		var rslt simResult
		rslt = <-ch
		if rslt.Done {
			routinesDone++
			continue
		}
		avgStrategizerWait := float64(rslt.SumStrategizerWaits) / float64(rslt.NumStrategizers)
		avgNonStrategizerWait := float64(rslt.SumNonStrategizerWaits) / float64(rslt.NumNonStrategizers)
		avgWait := float64(rslt.SumStrategizerWaits+rslt.SumNonStrategizerWaits) / float64(rslt.NumStrategizers+rslt.NumNonStrategizers)
		fmt.Printf("%0.2f,%0.2f,%0.2f,%02.f\n", rslt.PStrategy, avgStrategizerWait/1000.0, avgNonStrategizerWait/1000.0, avgWait/1000.0)
	}
}

func main() {
	SimPortaPotty()
}
