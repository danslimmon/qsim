package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/danslimmon/qsim"
)

type BloodBankArrProc struct {
	Sys *BloodBankSystem

	lastDraw int

	// Callback lists
	cbBeforeArrive []func(ap qsim.ArrProc)
	cbAfterArrive  []func(ap qsim.ArrProc, jobs []*qsim.Job, interval int)
}

// Arrive simulates the process of drawing new blood for the bank.
//
// We draw enough blood to fill the bank to its MaxOccupancy, unless we've already
// drawn as much as we can safely draw for the day.
func (arrProc *BloodBankArrProc) Arrive(clock int) (jobs []*qsim.Job, interval int) {
	arrProc.beforeArrive()
	sys := arrProc.Sys
	if clock-arrProc.lastDraw >= 1440 {
		var numToAppend int
		numToAppend = sys.MaxOccupancy - sys.queue.Length()
		if numToAppend > sys.MaxDrawRate*((clock-sys.lastDraw)/1440) {
			numToAppend = sys.MaxDrawRate * ((clock - sys.lastDraw) / 1440)
		}
		for i := 0; i < numToAppend; i++ {
			jobs = append(jobs, qsim.NewJob(clock))
		}
		sys.lastDraw = clock
	}
	arrProc.afterArrive(jobs, 1440)
	return jobs, 1440
}

func (arrProc *BloodBankArrProc) BeforeArrive(f func(ap qsim.ArrProc)) {
	arrProc.cbBeforeArrive = append(arrProc.cbBeforeArrive, f)
}
func (arrProc *BloodBankArrProc) beforeArrive() {
	for _, cb := range arrProc.cbBeforeArrive {
		cb(arrProc)
	}
}
func (arrProc *BloodBankArrProc) AfterArrive(f func(ap qsim.ArrProc, jobs []*qsim.Job, interval int)) {
	arrProc.cbAfterArrive = append(arrProc.cbAfterArrive, f)
}
func (arrProc *BloodBankArrProc) afterArrive(jobs []*qsim.Job, interval int) {
	for _, cb := range arrProc.cbAfterArrive {
		cb(arrProc, jobs, interval)
	}
}

type BloodBankSystem struct {
	// A slice of age thresholds for which we'll track stats.
	Thresholds []int
	// When to start capturing stats. We use this to avoid sampling the
	// initial ramp-up of the system.
	StatsStart int
	// The oldest age that units (jobs) are allowed to reach
	MaxJobAge int
	// The maximum draw rate (in units/day)
	MaxDrawRate int
	// The target (max) occupancy (in units)
	MaxOccupancy int
	// The average rate at which transfusions happen (in units/day)
	MeanTransfusionRate float64

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
	// The number of transfusions that had to be aborted due to a blood shortfall
	NumAborted int
	// The list of all ages of units used in transfusions (in ticks)
	UnitAges []int
	// For each age in Thresholds, the number of samples used that were older than
	// that age.
	AgeCounts []int

	statsStarted bool
	unitsUsed    []*qsim.Job
	lastDraw     int
}

// Init runs before the simulation begins, and its job is to set up the
// queues, processors, and behaviors.
//
// Ticks represent minutes.
func (sys *BloodBankSystem) Init() {
	var procMean float64

	rand.Seed(time.Now().UnixNano())
	// MeanTransfusionRate is in units/day, so the mean time between transfusions is
	// the reciprocal of that, expressed in ticks/unit
	procMean = 1440.0 / sys.MeanTransfusionRate
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
	sys.transfusionProcessor.AfterStart(func(p *qsim.Processor, j *qsim.Job, procTime int) {
		if sys.statsStarted && j != nil && j.ArrTime != -1 {
			sys.NumUsed++
			sys.unitsUsed = append(sys.unitsUsed, j)
		}
	})
	sys.trashProcessor.AfterFinish(func(p *qsim.Processor, j *qsim.Job) {
		if sys.statsStarted && j != nil {
			sys.NumTossed++
		}
	})

	sys.arrProc = &BloodBankArrProc{Sys: sys}
	sys.arrBeh = qsim.NewAlwaysQueueArrBeh(sys.queue, sys.arrProc)

	applyBloodBankDiscipline(sys, sys.queue, sys.trashProcessor, sys.transfusionProcessor)
	sys.transfusionProcessor.Start(nil)
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

// BeforeRun runs right before the clock starts.
func (sys *BloodBankSystem) BeforeFirstTick() {
	sys.transfusionProcessor.Start(qsim.NewJob(0))
}

// BeforeEvents runs at every tick when a simulation event happens (a
// Job arrives in the system, or a Job finishes processing and leaves
// the system). BeforeEvents is called after all the events for the tick
// in question have finished.
//
// In this example, we use BeforeEvents to send any jobs older than the
// maximum age to the trash.
func (sys *BloodBankSystem) BeforeEvents(clock int) {
	for _, j := range sys.queue.Jobs {
		if clock-j.ArrTime >= sys.MaxJobAge {
			sys.queue.Remove(j)
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
			sys.UnitAges = append(sys.UnitAges, clock-j.ArrTime)
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
	// Assigns the youngest unit to the processor
	assigner := func(cbProc *qsim.Processor, cbJob *qsim.Job) {
		var i, iYoungest int
		var j *qsim.Job
		if queue.Length() == 0 {
			qsim.D("Aborted after Job", cbJob)
			if sys.statsStarted {
				sys.NumAborted++
			}
			// Give the processor a dummy job until the next transfusion
			j = qsim.NewJob(-1)
			cbProc.Start(j)
			return
		}
		for i, j = range queue.Jobs {
			if j.ArrTime > queue.Jobs[iYoungest].ArrTime {
				iYoungest = i
			}
		}
		j = queue.Jobs[iYoungest]
		queue.Remove(j)
		cbProc.Start(j)
		qsim.D("Started Job", j)
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
//
// Arguments, in order:
//
// - Daily Maximum Draw Rate (in units/day, as an int)
// - Maximum Bank Occupancy (in units, as an int)
// - Daily Mean Transfusion Rate (in units/day, as a float)
// - (optional) The word "test", indicating that only 1 sim should be run.
//
// We output a CSV row with the following values, in order:
//
// - Total number of ticks ("minutes") for which simulations collected data
// - Number of simulations
// - Number of units used in transfusions
// - Number of units thrown out
// - Number of transfusions aborted due to lack of blood
// - 90th percentile age of units used in transfusions
// - A value for each member of `thresholds` (see below) indicating the number of units
//   used in transfusions over that age in days
func SimBloodBank() {
	var simTicks, nSims, simsPerCpu, statsStart int
	var maxDrawRate, maxOccupancy int
	var maxDrawRate64, maxOccupancy64 int64
	var meanTransfusionRate float64
	type simResult struct {
		Done                           bool
		NumTossed, NumUsed, NumAborted int
		UnitAges                       []int
		AgeCounts                      []int
	}
	var nTossed, nUsed, nAborted int
	var unitAges, ageCounts, thresholds []int
	var ch chan simResult
	var cpu, nCpu, routinesDone int
	var err error

	if len(os.Args) > 4 && os.Args[4] == "test" {
		nCpu = 1
		nSims = 1
		simTicks = 2 * 365 * 1440
	} else {
		nCpu = 16
		nSims = 64
		simTicks = 40 * 365 * 1440
	}
	simsPerCpu = nSims / nCpu
	// Don't start collecting stats until a year goes by
	statsStart = 365 * 1440

	thresholds = []int{5 * 1440, 10 * 1440, 15 * 1440, 20 * 1440, 25 * 1440, 30 * 1440}
	unitAges = make([]int, 0)
	ageCounts = make([]int, len(thresholds))

	maxDrawRate64, err = strconv.ParseInt(os.Args[1], 10, 0)
	if err != nil {
		panic("Failed to parse mean draw rate")
	}
	maxDrawRate = int(maxDrawRate64)
	maxOccupancy64, err = strconv.ParseInt(os.Args[2], 10, 0)
	if err != nil {
		panic("Failed to parse maximum occupancy")
	}
	maxOccupancy = int(maxOccupancy64)
	meanTransfusionRate, err = strconv.ParseFloat(os.Args[3], 0)
	if err != nil {
		panic("Failed to parse mean transfusion rate")
	}

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
						Thresholds:          thresholds,
						StatsStart:          statsStart,
						MaxDrawRate:         maxDrawRate,
						MaxOccupancy:        maxOccupancy,
						MeanTransfusionRate: meanTransfusionRate,
					}
					qsim.RunSimulation(sys, simTicks)

					rslt.NumTossed += sys.NumTossed
					rslt.NumUsed += sys.NumUsed
					rslt.NumAborted += sys.NumAborted
					rslt.UnitAges = append(rslt.UnitAges, sys.UnitAges...)
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
		nAborted += rslt.NumAborted
		unitAges = append(unitAges, rslt.UnitAges...)
		for i, _ := range thresholds {
			ageCounts[i] += rslt.AgeCounts[i]
		}
	}

	sort.Ints(unitAges)
	ind := int(9 * len(unitAges) / 10)
	p90UnitAge := unitAges[ind]

	fmt.Printf("%d,%d,%d,%d,%d,%d",
		(simTicks-statsStart)*nSims,
		nSims,
		nUsed,
		nTossed,
		nAborted,
		p90UnitAge,
	)
	for i, _ := range thresholds {
		fmt.Printf(",%d", ageCounts[i])
	}
	fmt.Printf("\n")
}

func main() {
	qsim.Debug = false
	SimBloodBank()
}
