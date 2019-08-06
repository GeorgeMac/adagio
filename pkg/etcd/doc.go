// package etcd contains types which enable etcd as a repository backend for
// the adagio workflow engine.
//
// Keyspace Design (etcd internals)
//
// Namespaces:
// v0/runs/   : runs namespace
// v0/states/ : states namespace
//
// Objects:
// v0/runs/<run-id>                              : Run{}
// v0/states/<state>/run/<run-id>/node/<node-id> : Node{}
//
// States: (waiting, ready, running, completed)
package etcd
