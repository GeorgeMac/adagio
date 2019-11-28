Adagio Architecture
-------------------

## Overview

Adagio is designed to facilitate workflow execution across a fleet of agents via an API known as the control plane.

![adagion architecture diagram](./architecture.svg)

### Workflow

> What is a workflow?

A workflow is a set of instructions for work defined in a graph structure. In particular a directed acyclic-graph. Meaning there are no cycles between nodes (vertices).
This ensures a topological sort can be performed on the workflow in order to plan and execute the order of work.

The workflow graph is serialized and sent to the control plane to be executed. The definitions for the graph model are defined using protobuf and can be
found within the `pkg/adagio` Go package in this project.
Each node (vertex) in the graph can be thought of as a single function call. The function has a signature defined within the specification of a node along
with arguments encoded as metadata.

Alongside this; a node can be further defined with automatic recovery instructions in the form of a number of retries for a specific error or failure condition.
This allows for recovery to be performed automatically by the agents deployed.

### Node

> What is a node?

A node at its core is a specification of work. The node definition literally contains a type called the node spec as its first property.
This node specification contains the function to be called, the metadata associated (arguments) and any retry specifications.
Alongside this specification there are some runtime properties which record the result conclusion and output of each execution attempt,
the creation and finish times, along with any inputs fed in from dependent nodes.

### Agent

> What is an agent?

An agent is a single thread of execution which attempts to claim, execute and report on a single node at a time.
One process could spawn multiple agents e.g. in separate go routines. This can be achieved using the `pkg/agent.Pool` with
an agent count of two or more.

When a workflow is instantiated (started) it becomes known as a *run*. Workers are notified of the nodes within the workflow run.
In particular, they are notified of the ready nodes. Initiality these are the nodes within the graph with no inbound edges.

```
(ready) --> (not ready)
            /
(ready) ---
```

A node becomes ready once all nodes which feed into it are completed with a successful conclusion.
Agents consume nodes, not workflow runs. This allows for execution of a workflow to be distributed across multiple agents.

An agent will only claim nodes which it can execute. This is decided based on the nodes specification runtime property.
If the agent has a function definition associated for the runtime it can execute it. Given this is the case it attempts to make a *claim*.

Only one agent can successfuly make a claim for a node. This is how we ensure nodes are executed by only one worker at a time.
Once the work has been performed the result is recorded, the graph state is updated in the database and any _newly ready_ nodes
are communicated with listening agents.

When an agent holds a claim on a node it must reported back a heartbeat while it performs execution.
A failure to do so will cause all agents to be notified of an orphaned node. An orphaned node will
be reclaimed by another agent and the result of execution be recorded as an error.
Given the node specifies a number of retries on the `error` condition the node may be moved back into the
ready state to be claimed again. This is how fail-over can be configured in the event an agent becomes "unhealthy".

> We recommend functions be idempotent in order to be retried safely.

Agents are built to facilitate your workflow needs and are the concern of operators and function providers.

### Control Plane API

> What is the control plane API?

The control plane is the entry point to get work done. It exposes a number of API actions to instantiate (start) workflows as runs, list existing runs
in the system, list available agents (for introspection purposes), inspect individual runs or just to get high level statistics on the overall counts
of runs, nodes and their states within the cluster.

It is exposed as a gRPC API, however, there is a prebuilt json API gateway which can be deployed and communicated with the gRPC one. This API comes
with an accompanying swagger specification. This has been used to generate the javascript client used within the adagio UI.

The control plane API is designed to serve your cluster consumers who need to execute workflows.
It needs to be operated, but is intended to be simple to deploy and monitor.

## Deployment

```
+------------------------------------------------------------------------------------+
|                                                                                    |
|                                                           scale horizontally       |
|                                                                                    |
|  +-----------+  +-----------+  +-----------+  +-----------+               +-----+  |
|  |           |  |           |  |           |  |           |               |     |  |
|  |   agent   |  |   agent   |  |   agent   |  |   agent   |               |  +  |  |
|  |           |  |           |  |           |  |           |      ...      |     |  |
|  +-----------+  +-----------+  +-----------+  +-----------+               +-----+  |
|        |              |              |              |                              |
+------------------------------------------------------------------------------------+
         |              |              |              |
         +--------------------------^----------------------------------------+ ...
                                    |            |
                                    |            |
                                    |            |
                             +-------------------v-------+    +-----------------+
                             |                           |    |                 |
                             |                           |    |  ✓ etcd         |
                             |  Multi-Row Transactional  |    |  - dynamoDB     |
                             |          Database         |    |  - PostgresSQL  |
                             |                           |    |  ...            |
                             |                           |    |                 |
                             +------^--------------------+    +-----------------+
                                    |            |
                                    |            |
         +---------------------------------------v---------------------------+ ...
         |              |              |              |
+------------------------------------------------------------------------------------+
|        |              |              +              |                              |
|        |              |        control plane        |     scale horizontally       |
|        |              |              +              |                              |
|  +-----------+  +-----------+  +-----------+  +-----------+               +-----+  |
|  |           |  |           |  |           |  |           |               |     |  |
|  |    api    |  |    api    |  |    api    |  |    api    |               |  +  |  |
|  |           |  |           |  |           |  |           |      ...      |     |  |
|  +-----------+  +-----------+  +-----------+  +-----------+               +-----+  |
|                                                                                    |
+------------------------------------------------------------------------------------+
```

Adagio is designed to facilitate the scale of agents which perform work and the control plane api processes horizontally.
It doesn't attempt to dictate how or where you deploy your agents or the api tier. Rather it implements the pieces which
collaborate via some multi-row transactional database to ensure horizontal scale can be achieved.

`adagiod` today serves as both the control plane API and _an_ implementation of an adagio agent. This implementation has some
limited functionality which is expected to not be of much use to the adagio users. While work progresses on the pre-baked agent
within `adagiod` it should be made clear that it is a pre-configured binary wrapper for the `pkg/adagio` Go package. This package
is intended for consumers to import and build agents of their own in Go. Please see `pkg/adagio/doc.go` for more details on using
this package to bake your own adagio agent in Go. In the future the goal will be to enable other agent function implementations
in other languages. This may be through some IPC or network based API for example.
