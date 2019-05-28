adagiod - the adagio work agent
-------------------------------

`adagiod` is a binary which contains both the control plan API and the work consuming agent. It can simultaneously be run as both by ommitting either of the sub-commands (agent or api). Otherwise, it can be run as a standalone api or agent process. Note that the in-memory (-backend-type=memory) backend repository is pointless when adagiod is configured as seperate api and agent processes. Instead use a remote backend for deployment scenarios such as this (currently only "etcd" is supported here).

```
Usage: adagiod <api|agent> [OPTIONS]

The adagio workflow agent and control plane API

Options:
  -backend-type string
    	backend repository type ("memory"|"etcd") (default "memory")
  -config string
    	location of config toml file
  -etcd-addresses string
    	list of etcd node addresses (default "http://127.0.0.1:2379")
```

## Example

see [example toml](../../example/config.toml) for configuration file example.

Requirements:

- etcd node running at "http://127.0.0.1:2379"

From the route of the adagio github project do the following in one terminal session:

```
adagiod -config example/config.toml
```

You are now running a combined adagio agent and api. You can solely run the api or agent by specifying either `api` or `agent` after `adagiod`.

In another session you can now communicate with the control plane like so:

```
adagio runs ls
```

## Repository Backends

### Memory

This is an in-memory implementation of the adagio protocol. Which is mostly a demonstration of adagio. Note that killing this process leads to all state created during the daemons lifetime to be lost.

### Etcd

This is an etcd backed implementation of the adagio repository protocol. It is designed such that api and agent can be deployed seperately and that multiple agents can be deployed and scaled elastically.
As long as they all share access to the same etcd cluster. Work will be distributed amongst the agents ensuring at most once execution of node operations per operation attempt, per run.
