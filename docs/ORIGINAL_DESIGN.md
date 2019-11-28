Adagio - Workflow Engine
------------------------

A system where workflows defined as directed acyclic graphs are "executed" on a set of distributed "workers" or "nodes".

## Responsibilities and Components 

- requirement  (\*)
- stretch goal (+)
- uncertain    (?)

### Graph Definition (\*)

Adagio the project should have a concrete, versioned definitions (most likely in protobuf) for the graph representation. One which can be represented "on the wire" in either protobuf or JSON.
This definition becomes the currency of the adagion system. Graphs can be started which produces a run. Workers within the system consume and process nodes once they are ready, at most once, unless they are being restarted.

###  State Persistence (?)

Adagio may or may not give guarantees around state. There will exist state within a backing storage mechanism which helps to coordinate work. However, constraints around consistency of this state have yet to be decided.

### State Progression (\*)

Adagio backing repository implementations must uphold guarantees such as progression of state for nodes on outbound edges from the currently owned node. This will be enforced by a testing harness which all implementations must uphold in order to function as a backing store. This is where adagio can be viewed as a specification for a protocol of work as much as it is an implementation.

### Work Allocation (\*)

As with state progression, adagio must support the allocation of work to workers within the system. This is also a function of the adagion protocol which should be covered by the test harness. The protocol dictates that a backing repository by provide both a way to _attempt a claim_ for a node and _subscribe to events_ which notify as to when a node moves into a new state. Particularly the ready state in the case of claiming a new piece of work.

### Failure Recovery (+)

Ultimately the execution of work will fail for reasons which belong to the adagio system or the pool of workers. Failures which are not faults of the author of the adagio run. This a system classification of error, rather than a consumer error. In this scenario the protocol and implementations should support detection of errors such as failure to report on claimed work or other system related errors. The it should be possible to re-allocate work elsewhere and make new attempts to process it. The protocol should support annotation of nodes for the consumer which describe suitable constraints around number of retries.

### Runtime Propagation (?)

Nodes are configured to be executed by an explicit runtime declaration. If a worker cannot process a node because it doesn't have the specified runtime, then it should not attempt to claim it.
How runtimes themselves are configured, installed and reported on at a entire pool level is not yet decided or explored. However, this may be an area worth considering in the future.

## The Interface

The system consists (roughly) of the following interface 

```
Start(Graph) (Run, error)
Inspect(id string) (Run, error)
List() ([]Run, error)
Stop(id string) error
```

Graph:       Is defined as a DAG with a single entrypoint vertex 
Start(Graph): Loads the workflow definition into the system which workers subscribe to

Vertices can be in one of a few states:

- waiting
- ready
- executing
- completed

All entrypoint vertices are immediately in a `ready` state.
Any downstream vertex (not the entrypoint) initializes in a `waiting` state.

Workers subscribe to `ready` vertices.
Workers make claims to execute `ready` vertices.
When a claim is made they progress the vertex into an `executing` state.
When a worker finishes execution of a vertex it progresses it into a `completed` state.
Along with _completing_ their currently claimed vertex, they also progress any dependent `waiting` vertices, given the current vertex is the last dependency. 
A vertex can be completed into one of many conclusions.

Vertices can be concluded as:

- successful
- failed
- unreached 
- cancelled
- orphaned (system failure)

When _failing_ a vertex, the node will mark downstream vertices as `unreached`.
When _succeeding_ a vertex, the node will progress outgoing vertices from `waiting` to `ready` (given these vertices aren't waiting on other different incoming vertices).

The entire workflow can be cancelled leading to `ready`, `waiting` and `executing` vertices to become `completed` with conclusion `cancelled`.

## Components

### Repository

Repositories are at the heart of adagio. Implementing adapters of this interface must uphold a contract which the adagio agent and control plane API depend upon.

```go
type Repository interface {
  Claim(Node) error
  Subscribe(chan<- Events) error
}
```

A test harness should be developed which can probe this interface and assert expecations for any implementation. The harness should enforce constraints such as one claim per node under concurrency access scenarios. As well as constraints around when nodes are in a state to be claimed (ready rather than waiting or executing).

## Runtime Ideas

- Simple language based runtime
- Docker based runtime
- Child process runtime
- Lambda-like runtime

## Bucket Of Left Over Thoughts

- Typed inputs and outputs for nodes
- Worked example of using adagio for continuous integration
