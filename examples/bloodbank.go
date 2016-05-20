package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/danslimmon/qsim"
)

type BloodBankSystem struct {
	// A slice of age thresholds for which we'll track stats.
	Thresholds []int

	// When to start capturing stats. We use this to avoid sampling the
	// initial ramp-up of the system.
	StatsStart int

	// The oldest age that units (jobs) are allowed to reach
	MaxJobAge int

	// The list of all queues in the system.
	queue *qsim.Queue
	// The list of all processors in the system.
	trashProcessor, transfusionProcessor *qsim.Processor
	// The system's arrival process
	arrProc qsim.ArrProc
	// The system's arrival behavior
	arrBeh qsim.ArrBeh

	// The total number of units that went through the trash can processor and the
	// transfusion processor, respectively.
	NumTossed, NumUsed int
	// For each age in Thresholds, the number of samples used that were older than
	// that age.
	AgeCounts []int

	statsStarted bool
	unitsUsed    []*qsim.Job
}

// Init runs before the simulation begins, and its job is to set up the
// queues, processors, and behaviors.
//
// Ticks represent minutes.
func (sys *BloodBankSystem) Init() {
	var arrivalMean, procMean float64

	rand.Seed(time.Now().UnixNano())
	arrivalMean = 480.0
	procMean = 720.0
	sys.MaxJobAge = 35 * 1440

	sys.AgeCounts = make([]int, len(sys.Thresholds))
	sys.unitsUsed = make([]*qsim.Job, 0)

	transfusionIntervalGenerator := func(j *qsim.Job) int {
		var r float64
		r = rand.ExpFloat64() * procMean
		return int(r)
	}

	// There is only one queue, representing the fridge.
	sys.queue = qsim.NewQueue()

	// There are two processors. One represents the trash can, and one represents
	// transfusions.
	sys.trashProcessor = qsim.NewProcessor(func(j *qsim.Job) int { return 0 })
	sys.transfusionProcessor = qsim.NewProcessor(transfusionIntervalGenerator)

	// Processor callbacks to keep track of stats.
	sys.transfusionProcessor.AfterFinish(func(p *qsim.Processor, j *qsim.Job) {
		if sys.statsStarted {
			sys.NumUsed++
			sys.unitsUsed = append(sys.unitsUsed, j)
		}
	})
	sys.trashProcessor.AfterFinish(func(p *qsim.Processor, j *qsim.Job) {
		if sys.statsStarted {
			sys.NumTossed++
		}
	})

	sys.arrProc = qsim.NewPoissonArrProc(arrivalMean)

	// We _always_ want to queue, and there's only one queue, so this is a
	// degenerate case of ShortestQueueArrBeh.
	sys.arrBeh = qsim.NewShortestQueueArrBeh(
		[]*qsim.Queue{sys.queue},
		[]*qsim.Processor{sys.transfusionProcessor},
	)

	applyBloodBankDiscipline(sys, sys.queue, sys.trashProcessor, sys.transfusionProcessor)
}

// ArrProc returns the system's arrival process.
func (sys *BloodBankSystem) ArrProc() qsim.ArrProc {
	return sys.arrProc
}

// ArrBeh returns the system's arrival behavior.
func (sys *BloodBankSystem) ArrBeh() qsim.ArrBeh {
	return sys.arrBeh
}

// Processors returns the list of Processors in the system.
func (sys *BloodBankSystem) Processors() []*qsim.Processor {
	return []*qsim.Processor{sys.transfusionProcessor, sys.trashProcessor}
}

// BeforeEvents runs at every tick when a simulation event happens (a
// Job arrives in the system, or a Job finishes processing and leaves
// the system). BeforeEvents is called after all the events for the tick
// in question have finished.
func (sys *BloodBankSystem) BeforeEvents(clock int) {
	for i, j := range sys.queue.Jobs {
		if clock-j.ArrTime >= sys.MaxJobAge {
			sys.queue.Jobs = append(sys.queue.Jobs[:i], sys.queue.Jobs[i+1:]...)
			sys.trashProcessor.Start(j)
		}
	}
}

// AfterEvents runs at every tick when a simulation event happens, but
// in contrast with BeforeEvents, it runs after all the events for that
// tick have occurred.
func (sys *BloodBankSystem) AfterEvents(clock int) {
	if clock >= sys.StatsStart {
		sys.statsStarted = true
	}
	if sys.statsStarted {
		for _, j := range sys.unitsUsed {
			for i, thresh := range sys.Thresholds {
				if clock-j.ArrTime > thresh {
					sys.AgeCounts[i]++
				}
			}
		}
		sys.unitsUsed = sys.unitsUsed[:0]
	}
}

func applyBloodBankDiscipline(sys *BloodBankSystem, queue *qsim.Queue, trashProcessor, transfusionProcessor *qsim.Processor) {
	/*
		// Assigns a random job from the queue to the processor
		assigner := func(cbProc *qsim.Processor, cbJob *qsim.Job) {
			var i int
			var j *qsim.Job
			if queue.Length() == 0 {
				return
			}
			i = rand.Intn(queue.Length())
			j = queue.Jobs[i]
			queue.Jobs = append(queue.Jobs[:i], queue.Jobs[i+1:]...)
			cbProc.Start(j)
		}
	*/

	// Assigns the youngest unit to the processor
	assigner := func(cbProc *qsim.Processor, cbJob *qsim.Job) {
		var i, iYoungest int
		var j *qsim.Job
		if queue.Length() == 0 {
			return
		}
		for i, j = range queue.Jobs {
			if j.ArrTime > queue.Jobs[iYoungest].ArrTime {
				iYoungest = i
			}
		}
		j = queue.Jobs[iYoungest]
		queue.Jobs = append(queue.Jobs[:iYoungest], queue.Jobs[iYoungest+1:]...)
		cbProc.Start(j)
	}

	transfusionProcessor.AfterFinish(assigner)
}

// Simulates a blood bank.
//
// – Any blood unit older than 35 days is thrown in the trash.
// – There is only one queue, representing the bank itself.
// - There are two processors. One represents the trash, and the other
//   represents actual use in transfusion.
// – Each tick is a minute.
func SimBloodBank() {
	var simTicks, nSims, simsPerCpu int
	type simResult struct {
		Done               bool
		NumTossed, NumUsed int
		AgeCounts          []int
	}
	var nTossed, nUsed int
	var ageCounts []int
	var thresholds []int
	var ch chan simResult
	var cpu, nCpu, routinesDone int

	nCpu = 1
	nSims = 1
	simsPerCpu = nSims / nCpu
	// Run each simulation for 10 years
	simTicks = 10 * 365 * 1440

	thresholds = []int{5 * 1440, 10 * 1440, 15 * 1440, 20 * 1440, 25 * 1440, 30 * 1440}
	ageCounts = make([]int, len(thresholds))

	ch = make(chan simResult)
	for cpu = 0; cpu < nCpu; cpu++ {
		go func(cpu int) {
			var i int
			for i = cpu*simsPerCpu + 1; i <= (cpu+1)*simsPerCpu; i++ {
				var rslt simResult
				rslt = simResult{
					AgeCounts: make([]int, len(thresholds)),
				}
				var j int
				for j = 0; j < simsPerCpu; j++ {
					sys := &BloodBankSystem{
						Thresholds: thresholds,
						StatsStart: 365 * 1440,
					}
					qsim.RunSimulation(sys, simTicks)

					rslt.NumTossed += sys.NumTossed
					rslt.NumUsed += sys.NumUsed
					var k int
					for k, _ = range rslt.AgeCounts {
						rslt.AgeCounts[k] = sys.AgeCounts[k]
					}
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
		nTossed += rslt.NumTossed
		nUsed += rslt.NumUsed
		for i, _ := range thresholds {
			ageCounts[i] += rslt.AgeCounts[i]
		}
	}

	fmt.Printf("Units tossed: %d\n", nTossed)
	fmt.Printf("Units used:   %d\n", nUsed)
	for i, thresh := range thresholds {
		fmt.Printf("Units used older than %2d days: %d\n", thresh/1440, ageCounts[i])
	}
}

func main() {
	SimBloodBank()
}
