Adagio - The Workflow Engine
----------------------------

> This project is currently in a constant state of flux. Don't expect it to work. Thank you o/

Adagio is a workflow execution tool designed to run both locally and facilitate execution across a cluster of worker nodes.
A workflow is a directed acyclic graph (DAG) within which the vertices describe the work to be executed. The inputs and outputs are carried along the edges of the graph, piping the result from one vertex to the next. Each vertex awaits the execution of all its inbound edges to finish. Given all inputs finish successfully the vertex will be allocated and executed.

Adagio focusses primarily on orchestration of workflow execution. The execution of "work" defined within the vertices of the graph is intended to be extensible via a combination of native handlers and a plugin architecture. Rather than dictated by what can be implemented within this project.

# Usage

## adagio

The adagio cli tool communicates with the control plane API.

```
adagio
adagio help          # show adagio command usage
adagio version       # show adagio version information

adagio runs          # adagio runs usage

adagio runs ls       # list runs
adagio runs start    # create and start runs
adagio runs inspect  # inspect a single run
adagio runs stop     # stop an active run
```

## adagiod 

Contains both the adagio control plane and the runtime agent. The can serve both at the same time. Otherwise, it can be deployed seperately controlled via configuration.

```
adagiod       # control plane + runtime agent 
adagiod api   # control plane
adagiod agent # runtime agent
```

# API Specification

The current API specifications can be found within the documentation for this project.

- [API V0](./docs/api/v0.md)
