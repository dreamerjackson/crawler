package spider

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type EventType int

const (
	EventTypeDelete EventType = iota
	EventTypePut

	RESOURCEPATH = "/resources"
)

type WatchResponse struct {
	Typ      EventType
	Res      *ResourceSpec
	Canceled bool
}

type WatchChan chan WatchResponse

type ResourceRegistry interface {
	GetResources() ([]*ResourceSpec, error)
	WatchResources() WatchChan
}

type EtcdRegistry struct {
	etcdCli *clientv3.Client
}

func NewEtcdRegistry(endpoints []string) (ResourceRegistry, error) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	return &EtcdRegistry{cli}, err
}

func (e *EtcdRegistry) GetResources() ([]*ResourceSpec, error) {
	resp, err := e.etcdCli.Get(context.Background(), RESOURCEPATH, clientv3.WithPrefix(), clientv3.WithSerializable())
	if err != nil {
		return nil, err
	}
	resources := make([]*ResourceSpec, 0)
	for _, kv := range resp.Kvs {
		r, err := Decode(kv.Value)
		if err == nil && r != nil {
			resources = append(resources, r)
		}
	}
	return resources, nil
}

func (e *EtcdRegistry) WatchResources() WatchChan {
	ch := make(WatchChan)
	go func() {
		watch := e.etcdCli.Watch(context.Background(), RESOURCEPATH, clientv3.WithPrefix(), clientv3.WithPrevKV())
		for w := range watch {
			if w.Err() != nil {
				zap.S().Error("watch resource failed", zap.Error(w.Err()))
				continue
			}
			if w.Canceled {
				zap.S().Error("watch resource canceled")
				ch <- WatchResponse{
					Canceled: true,
				}
			}
			for _, ev := range w.Events {

				switch ev.Type {
				case clientv3.EventTypePut:
					spec, err := Decode(ev.Kv.Value)
					if err != nil {
						zap.S().Error("decode etcd value failed", zap.Error(err))
					}
					if ev.IsCreate() {
						zap.S().Info("receive create resource", zap.Any("spec", spec))
					} else if ev.IsModify() {
						zap.S().Info("receive update resource", zap.Any("spec", spec))
					}

					ch <- WatchResponse{
						EventTypePut,
						spec,
						false,
					}

				case clientv3.EventTypeDelete:
					spec, err := Decode(ev.PrevKv.Value)
					zap.S().Info("receive delete resource", zap.Any("spec", spec))
					if err != nil {
						zap.S().Error("decode etcd value failed", zap.Error(err))
					}
					ch <- WatchResponse{
						EventTypeDelete,
						spec,
						false,
					}
				}
			}
		}
	}()

	return ch

}
