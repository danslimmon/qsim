package qsim

import (
	"math/rand"
	"sort"
)

// An ArrBeh ("arrival behavior") assigns new jobs to queues or processors.
type ArrBeh interface {
	Assign(j *Job) Assignment
	BeforeAssign(f func(ab ArrBeh, j *Job) *Assignment)
	AfterAssign(f func(ab ArrBeh, j *Job, ass Assignment))
}

// ShortestQueueArrBeh implements the ArrBeh interface, assigning new
// Jobs by the following algorithm:
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
	// Queues contains all the queues known to us.
	Queues []*Queue
	// IdleProcessors keeps track of which Processors are idle. A Processor
	// is a key in this map iff it is idle.
	IdleProcessors map[*Processor]bool

	// Callback lists
	cbBeforeAssign []func(ab ArrBeh, j *Job) *Assignment
	cbAfterAssign  []func(ab ArrBeh, j *Job, ass Assignment)
}

// Assign takes the given Job and assigns it to a queue or a processor.
// The documentation for ShortestQueueArrBeh describes the logic used in
// this implementation.
func (ab *ShortestQueueArrBeh) Assign(j *Job) Assignment {
	var proc *Processor
	var procs []*Processor
	var q *Queue
	var shortQueues []*Queue
	var i, smallestLength int
	var ass Assignment
	var assPtr *Assignment

	// Allow beforeAssign callback to override the assignment logic
	assPtr = ab.beforeAssign(j)
	if assPtr != nil {
		ab.assign(j, *assPtr)
		ab.afterAssign(j, *assPtr)
		return *assPtr
	}

	// Assign to an idle processor if there is at least one
	if len(ab.IdleProcessors) >= 1 {
		procs = make([]*Processor, 0)
		for proc, _ = range ab.IdleProcessors {
			procs = append(procs, proc)
		}

		i = rand.Intn(len(procs))
		ass = Assignment{Type: "Processor", Processor: procs[i]}
		ab.assign(j, ass)
		ab.afterAssign(j, ass)
		return ass
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
	ass = Assignment{Type: "Queue", Queue: q}
	ab.assign(j, ass)
	ab.afterAssign(j, ass)
	return ass
}

// assign does the appropriate thing with the Job given an Assignment.
func (ab *ShortestQueueArrBeh) assign(j *Job, ass Assignment) {
	switch ass.Type {
	case "Processor":
		ass.Processor.Start(j)
		D("Job", j.JobId, "arrived and was assigned to Processor", ass.Processor)
	case "Queue":
		ass.Queue.Append(j)
		D("Job", j.JobId, "arrived and was assigned to Queue", ass.Queue)
	default:
		panic("Tried to process Assignment with unknown Type '" + ass.Type + "'")
	}
}

// BeforeAssign adds a callback to run immediately before the Arrival Behavior
// assigns a job to a Queue or Processor. This callback is passed the ArrBeh
// itself as well as the Job that's about to be assigned.
//
// The callback may return an Assignment pointer. If it does so, this Assignment
// will override the ArrBeh's assignment logic. Otherwise, if the callback
// returns <nil>, the assignment will proceed normally.
//
// If there are multiple BeforeAssign callbacks that return non-nil Assignment
// pointers, the callback most recently created wins.
func (ab *ShortestQueueArrBeh) BeforeAssign(f func(ArrBeh, *Job) *Assignment) {
	ab.cbBeforeAssign = append(ab.cbBeforeAssign, f)
}
func (ab *ShortestQueueArrBeh) beforeAssign(j *Job) *Assignment {
	var assPtr, newAssPtr *Assignment
	for _, cb := range ab.cbBeforeAssign {
		newAssPtr = cb(ab, j)
		if newAssPtr != nil {
			assPtr = newAssPtr
		}
	}
	return assPtr
}

// BeforeAssign adds a callback to run immediately after the Arrival Behavior
// assigns a job to a Queue or Processor. This callback is passed the ArrBeh
// itself, the Job that's about to be assigned, and an Assignment struct
// indicating where the Job was placed.
func (ab *ShortestQueueArrBeh) AfterAssign(f func(ArrBeh, *Job, Assignment)) {
	ab.cbAfterAssign = append(ab.cbAfterAssign, f)
}
func (ab *ShortestQueueArrBeh) afterAssign(j *Job, ass Assignment) {
	for _, cb := range ab.cbAfterAssign {
		cb(ab, j, ass)
	}
}

// NewShortestQueueArrBeh initializes a ShortestQueueArrBeh with the given Queues &
// Processors.
func NewShortestQueueArrBeh(queues []*Queue, procs []*Processor, ap ArrProc) ArrBeh {
	var ab *ShortestQueueArrBeh
	var p *Processor

	ab = new(ShortestQueueArrBeh)
	ab.Queues = queues
	ab.IdleProcessors = make(map[*Processor]bool)
	for _, p = range procs {
		if p.IsIdle() {
			ab.IdleProcessors[p] = true
		}
	}

	// These callbacks keep ab.IdleProcessors up to date.
	afterStart := func(p *Processor, j *Job, procTime int) {
		delete(ab.IdleProcessors, p)
	}
	afterFinish := func(p *Processor, j *Job) {
		ab.IdleProcessors[p] = true
	}
	for _, p = range procs {
		p.AfterStart(afterStart)
		p.AfterFinish(afterFinish)
	}

	// Make sure that newly arriving Jobs get assigned.
	ap.AfterArrive(func(cbArrProc ArrProc, cbJobs []*Job, cbInterval int) {
		for _, j := range cbJobs {
			ab.Assign(j)
		}
	})

	return ab
}

// AlwaysQueueArrBeh always puts incoming jobs in the given queue. Processors
// don't even enter into it.
type AlwaysQueueArrBeh struct {
	Q *Queue

	// Callback lists
	cbBeforeAssign []func(ab ArrBeh, j *Job) *Assignment
	cbAfterAssign  []func(ab ArrBeh, j *Job, ass Assignment)
}

// Assign takes the given Job and assigns it to the queue.
func (ab *AlwaysQueueArrBeh) Assign(j *Job) Assignment {
	// Allow beforeAssign callback to override the assignment logic
	assPtr := ab.beforeAssign(j)
	if assPtr != nil {
		ab.assign(j, *assPtr)
		ab.afterAssign(j, *assPtr)
		return *assPtr
	}

	ass := Assignment{Type: "Queue", Queue: ab.Q}
	ab.assign(j, ass)
	ab.afterAssign(j, ass)
	return ass
}

// assign does the appropriate thing with the Job given an Assignment.
func (ab *AlwaysQueueArrBeh) assign(j *Job, ass Assignment) {
	switch ass.Type {
	case "Processor":
		panic("AlwaysQueueArrBeh does not support assignment to Processors")
	case "Queue":
		ass.Queue.Append(j)
		D("Job", j.JobId, "arrived and was assigned to Queue", ass.Queue)
	default:
		panic("Tried to process Assignment with unknown Type '" + ass.Type + "'")
	}
}

// BeforeAssign adds a callback to run immediately before the Arrival Behavior
// assigns a job to a Queue or Processor. This callback is passed the ArrBeh
// itself as well as the Job that's about to be assigned.
//
// The callback may return an Assignment pointer. If it does so, this Assignment
// will override the ArrBeh's assignment logic. Otherwise, if the callback
// returns <nil>, the assignment will proceed normally.
//
// If there are multiple BeforeAssign callbacks that return non-nil Assignment
// pointers, the callback most recently created wins.
func (ab *AlwaysQueueArrBeh) BeforeAssign(f func(ArrBeh, *Job) *Assignment) {
	ab.cbBeforeAssign = append(ab.cbBeforeAssign, f)
}
func (ab *AlwaysQueueArrBeh) beforeAssign(j *Job) *Assignment {
	var assPtr, newAssPtr *Assignment
	for _, cb := range ab.cbBeforeAssign {
		newAssPtr = cb(ab, j)
		if newAssPtr != nil {
			assPtr = newAssPtr
		}
	}
	return assPtr
}

// AfterAssign adds a callback to run immediately after the Arrival Behavior
// assigns a job to a Queue or Processor. This callback is passed the ArrBeh
// itself, the Job that's about to be assigned, and an Assignment struct
// indicating where the Job was placed.
func (ab *AlwaysQueueArrBeh) AfterAssign(f func(ArrBeh, *Job, Assignment)) {
	ab.cbAfterAssign = append(ab.cbAfterAssign, f)
}
func (ab *AlwaysQueueArrBeh) afterAssign(j *Job, ass Assignment) {
	for _, cb := range ab.cbAfterAssign {
		cb(ab, j, ass)
	}
}

// NewAlwaysQueueArrBeh initializes a AlwaysQueueArrBeh with the given Queue.
func NewAlwaysQueueArrBeh(q *Queue, ap ArrProc) ArrBeh {
	var ab *AlwaysQueueArrBeh

	ab = new(AlwaysQueueArrBeh)
	ab.Q = q

	// Make sure that newly arriving Jobs get assigned.
	ap.AfterArrive(func(cbArrProc ArrProc, cbJobs []*Job, cbInterval int) {
		for _, j := range cbJobs {
			ab.Assign(j)
		}
	})

	return ab
}

// An Assignment indicates where a Job has been assigned by an Arrival Behavior.
//
// The string Type will be either "Processor" or "Queue", and the corresponding
// attribute (either Processor or Queue) will contain the entity to which the
// Job was assigned. The other attribute will be nil.
type Assignment struct {
	Type      string
	Processor *Processor
	Queue     *Queue
}

// ByQueueLength implements sort.Interface for []*Queue based on Length.
type ByQueueLength []*Queue

func (a ByQueueLength) Len() int           { return len(a) }
func (a ByQueueLength) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByQueueLength) Less(i, j int) bool { return a[i].Length() < a[j].Length() }
