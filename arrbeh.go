package qsim

import (
	"math/rand"
	"sort"
)

// An ArrBeh ("arrival behavior") assigns new jobs to queues or processors.
type ArrBeh interface {
	// Assign takes the given Job and assigns it (according to the
	// implementation) to a queue or a procesor.
	Assign(j *Job)
	// BeforeAssign adds a callback to be run immediately before a Job is
	// assigned to a Queue or Processor.
	BeforeAssign(f func(ab ArrBeh, j *Job))
}

// ShortestQueueArrBeh assigns new Jobs by the following algorithm:
//
// – If there is at least one idle Processor, pick an idle Processor at
//   random and start the Job on it.
// – Otherwise, append the Job to the shortest Queue available. If the
//   shortest queue length is shared by more than one Queue, the Job is
//   appended to one of those Queues at random.
//
// This behavior is like that of a supermarket checkout line: if there's
// an empty aisle you go straight there; otherwise you find the shortest
// queue and join it.
type ShortestQueueArrBeh struct {
	// queues contains all the queues known to us.
	Queues []*Queue
	// processors keeps track of which Processors are idle. A Processor
	// is a key in this map iff it is idle.
	Processors map[*Processor]bool

	// Callback lists
	cbBeforeAssign []func(ab ArrBeh, j *Job)
	cbAfterAssign  []func(ab ArrBeh, j *Job)
}

// See the documentation for ShortestQueueArrBeh.
func (ab *ShortestQueueArrBeh) Assign(j *Job) {
	var proc *Processor
	var procs []*Processor
	var q *Queue
	var shortQueues []*Queue
	var i, smallestLength int

	if len(ab.Processors) >= 1 {
		procs = make([]*Processor, 0)
		for proc, _ = range ab.Processors {
			procs = append(procs, proc)
		}

		ab.beforeAssign(j)
		if len(procs) == 1 {
			procs[0].Start(j)
		} else {
			// There is more than one idle Processor, so we have to pick one at random.
			i = rand.Intn(len(procs))
			procs[i].Start(j)
		}
		return
	}

	// If we've arrived here, then there are no idle Processors.
	sort.Sort(ByQueueLength(ab.Queues))
	smallestLength = ab.Queues[0].Length()
	shortQueues = make([]*Queue, 0, len(ab.Queues))
	for i = 0; i < len(ab.Queues) && ab.Queues[i].Length() == smallestLength; i++ {
		shortQueues = append(shortQueues, ab.Queues[i])
	}
	// Pick a random element from the list of queues that have the shortest length.
	i = rand.Intn(len(shortQueues))
	q = shortQueues[i]
	ab.beforeAssign(j)
	q.Append(j)
	return
}

func (ab *ShortestQueueArrBeh) BeforeAssign(f func(ArrBeh, *Job)) {
	ab.cbBeforeAssign = append(ab.cbBeforeAssign, f)
}
func (ab *ShortestQueueArrBeh) beforeAssign(j *Job) {
	for _, cb := range ab.cbBeforeAssign {
		cb(ab, j)
	}
}

// NewShortestQueueArrBeh initializes a ShortestQueueArrBeh with the given Queues &
// Processors.
func NewShortestQueueArrBeh(queues []*Queue, procs []*Processor) ArrBeh {
	var ab *ShortestQueueArrBeh
	var p *Processor

	ab = new(ShortestQueueArrBeh)
	ab.Queues = queues
	ab.Processors = make(map[*Processor]bool)
	for _, p = range procs {
		if p.IsIdle() {
			ab.Processors[p] = true
		}
	}

	// These callbacks keep ab.Processors up to date.
	afterStart := func(p *Processor, j *Job, procTime int) {
		delete(ab.Processors, p)
	}
	afterFinish := func(p *Processor, j *Job) {
		ab.Processors[p] = true
	}
	for _, p = range procs {
		p.AfterStart(afterStart)
		p.AfterFinish(afterFinish)
	}

	return ab
}

// ByQueueLength implements sort.Interface for []*Queue based on Length.
type ByQueueLength []*Queue

func (a ByQueueLength) Len() int           { return len(a) }
func (a ByQueueLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByQueueLength) Less(i, j int) bool { return a[i].Length() < a[j].Length() }
