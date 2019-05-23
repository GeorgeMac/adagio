package etcd

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgemac/adagio/pkg/adagio"
	"github.com/gogo/protobuf/proto"
	"go.etcd.io/etcd/clientv3"
)

type Repository struct {
	kv clientv3.KV
}

func New(kv clientv3.KV) *Repository {
	return &Repository{kv}
}

func (r *Repository) StartRun(spec *adagio.GraphSpec) (run *adagio.Run, err error) {
	run, err = adagio.NewRun(spec)
	if err != nil {
		return
	}

	data, err := proto.Marshal(run)
	if err != nil {
		return
	}

	var (
		runKey = runKey(run)
		cmps   = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(runKey), "=", 0),
		}
		ops = []clientv3.Op{
			clientv3.OpPut(runKey, string(data)),
		}
	)

	for _, node := range run.Nodes {
		nodeData, err := proto.Marshal(node)
		if err != nil {
			return nil, err
		}

		if node.State == adagio.Node_READY {
			put := clientv3.OpPut(nodeReadyKey(run, node), string(nodeData))
			ops = append(ops, put)
		}
	}

	resp, err := r.kv.Txn(context.Background()).
		If(cmps...).
		Then(ops...).
		Commit()
	if err != nil {
		return
	}

	if !resp.Succeeded {
		err = errors.New("duplicate run already created")
	}

	return
}

func runKey(run *adagio.Run) string {
	return fmt.Sprintf("runs/%s", run.Id)
}

func nodeReadyKey(run *adagio.Run, node *adagio.Node) string {
	return fmt.Sprintf("ready/run/%s/node/%s", run.Id, node.Spec.Name)
}

func (r *Repository) ListRuns() (runs []*adagio.Run, err error) {
	resp, err := r.kv.Get(context.Background(), "runs/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, kv := range resp.Kvs {
		var run adagio.Run

		if err = proto.Unmarshal([]byte(kv.Value), &run); err != nil {
			return
		}

		runs = append(runs, &run)
	}

	return
}

func (r *Repository) ClaimNode(run *adagio.Run, name string) (*adagio.Node, bool, error) {
	panic("not implemented")
}

func (r *Repository) FinishNode(run *adagio.Run, name string) error {
	panic("not implemented")
}

func (r *Repository) Subscribe(events chan<- *adagio.Event, states ...adagio.Node_State) error {
	panic("not implemented")
}
