Adagio - Workflow Engine
-----------------------------------

A system where workflows defined as directed acyclic graphs are "executed" on a set of distributed "workers" or "nodes".

## Responsibilities and Components 

requirement  *
stretch goal +
uncertain    ?

- Graph Definition     *
- State Persistence    ?
- State Progression    *
- Work Allocation      *
- Failure Recovery     +
- Runtime Propagation  ?

## The Interface

The system consists (roughly) of the following interface 

```
Exec(Graph) (Run, error)
Inspect(id string) (Run, error)
List() ([]Run, error)
Stop(id string) error
```

Graph:       Is defined as a DAG with a single entrypoint vertex 
Exec(Graph): Loads the workflow definition into the system which workers subscribe to

Vertices can be in multiple states:

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
