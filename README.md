Adagio - The Workflow Engine
----------------------------

> This project is currently in a constant state of flux. Don't expect it to work. Thank you o/

Adagio is a workflow execution tool designed to run both locally and facilitate execution across a cluster of worker nodes.
A workflow is a directed acyclic graph (DAG) within which the vertices describe the work to be executed. The inputs and outputs are carried along the edges of the graph, piping the result from one vertex to the next. Each vertex awaits the execution of all its inbound edges to finish. Given all inputs finish successfully the vertex will be allocated and executed.

Adagio focusses primarily on orchestration of workflow execution. The execution of "work" defined within the vertices of the graph is intended to be extensible via a combination of native handlers and a plugin architecture. Rather than dictated by what can be implemented within this project.

# Usage

## adagio - cli

The adagio cli tool communicates with the control plane API.

```
adagio
adagio help          # show adagio command usage

adagio runs          # adagio runs usage

adagio runs ls             # list runs
adagio runs start [file]   # create and start runs
adagio runs start <stdin>
```

## adagiod - service

Contains both the adagio control plane and the runtime agent. The can serve both at the same time. Otherwise, it can be deployed seperately controlled via configuration.

```
adagiod       # control plane + runtime agent 
adagiod api   # control plane
adagiod agent # runtime agent
```

### adagiod api

The `adagiod api` is a control plane daemon for an adagio system. It exposes the necessary interface for listing, inspecting and starting runs. It does so via the backing storage implementation.

The current specification is etched in protobuf. Client and Server code is generated using [Twirp](github.com/twitchtv/twirp).
The proto files and generated code is located within the [controlplane rpc](./pkg/rpc/controlplane) folder.

### adagiod agent

The `adagiod agent` is a daemon which consumes and processes nodes within graphs stored within a backing storage.

# Building


```
make help      # Print description of the commands available
make install   # Install adagio and adagiod
make protobuf  # Build protocol buffers into twirp model and service definitions
```

see `make help` for details locally.
