# qsim

`qsim` is a Go package that lets you build queueing system simulators.

The package provides some building blocks that you can customize and fit
together to simulate all kinds of queueing systems, from a grocery store
checkout line to a kanban board.

A queueing **system** in `qsim` processes arbitrary **jobs** and is
composed of 5 pieces:

* The **arrival process** controls how often jobs enter the system.
* The **arrival behavior** defines what happens when a new job arrives.
  When the arrival process generates a new job, the arrival behavior
  either sends it straight to a processor or appends it to a queue.
* **Queues** are simply holding pens for jobs. A system may have many
  queues associated with different processors.
* A **queueing discipline** defines the relationship between queues and
  processors. It's responsible for choosing the next job to process and
  assigning that job to a processor.
* **Processors** are the entities that remove jobs from the system.
  A processor may take differing amounts of time to process different
  jobs. Once a job has been processed, it leaves the queueing system.

To answer questions about a queueing system, we simulate its behavior
over a certain number of **ticks**. We can use **callbacks** to extract
the current system state at any point in the simulation and turn that
state into data.

## An example: supermarket checkout line.

Suppose you want to model the queueing behavior at a small supermarket
with 3 checkout lines. Here's the sort of queueing system you'd create
with `qsim`:

* **Arrival Process**: The arrival process is simple. A new job
  ("shopper") enters the queueing system ("becomes ready for checkout")
  every *n* seconds, where *n* is picked from some probability
  distribution you define.
* **Arrival Behavior**: When a job enters the system, it goes
  straight to any processor that is idle (i.e. any checkout lane that
  is empty). If there are no idle processors, the job enters the
  shortest queue available.
* **Queues**: There are 3 queues. At any time they each contain some
  number of jobs.
* **Queueing Discipline**: When a processor finishes a job (a cashier
  finishes checking a shopper out), the queueing discipline says that
  the next job in that processor's queue begins processing. Thus the
  queueing discipline is responsible for keeping track of the 1-to-1
  relationship between queues and processors.
* **Processor**: There are 3 processors, each of which represents a
  checkout lane. Each processor takes a certain time to process jobs
  ("checkout shoppers"), and that processing time is also drawn from
  a random distribution defined by you.

By putting these building blocks together, you can simulate supermarket
checkouts with shocking fidelity. By judiciously placing callbacks, you
can answer questions like:

1. Within how many seconds do 90% of shoppers complete the entire
   checkout process, from entering the queue to walking out the door?
2. What happens if a register needs to close for 20 minutes?
3. How much benefit could be gained from updating the scanners to
   more modern equipment that can scan items in fewer tries?

If you want to tweak the way the simulation works, all you have to do
is modify one of these building blocks.

For an example of this exact simulation, check out:

https://github.com/danslimmon/qsim/blob/master/grocery_test.go

## Another example: porta-potties

https://danslimmon.wordpress.com/2015/07/23/minding-your-pees-and-queues/
