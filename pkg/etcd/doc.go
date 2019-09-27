// package etcd contains types which enable etcd as a repository backend for
// the adagio workflow engine.
//
// Keyspace Design (etcd internals)
//
// Namespaces:
// v0/runs/   : runs namespace
// v0/nodes/  : nodes namespace
// v0/states/ : states namespace
//
// Objects:
// v0/agents/<agent-id>                       : Agent{} serialized agent object (leased)
// v0/runs/<run-id>                           : Run{}   serialized run object
// v0/nodes/<run-id>/node/<name>              : Node{}  serialized node object
// v0/states/<state>/run/<run-id>/node/<name> : ""      empty string to identify state
//
// States: waiting, ready, running, completed
package etcd
