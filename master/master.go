package master

import (
	"context"
	"errors"
	"fmt"
	"github.com/dreamerjackson/crawler/cmd/worker"
	"go-micro.dev/v4/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"net"
	"reflect"
	"sync/atomic"
	"time"
)

type Master struct {
	ID        string
	ready     int32
	leaderID  string
	workNodes map[string]*registry.Node
	options
}

func New(id string, opts ...Option) (*Master, error) {
	m := &Master{}

	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	m.options = options

	ipv4, err := getLocalIP()
	if err != nil {
		return nil, err
	}
	m.ID = genMasterID(id, ipv4, m.GRPCAddress)
	m.logger.Sugar().Debugln("master_id:", m.ID)
	go m.Campaign()

	return &Master{}, nil
}

func genMasterID(id string, ipv4 string, GRPCAddress string) string {
	return "master" + id + "-" + ipv4 + GRPCAddress
}

func (m *Master) IsLeader() bool {
	return atomic.LoadInt32(&m.ready) != 0
}

func (m *Master) Campaign() {
	endpoints := []string{m.registryURL}
	cli, err := clientv3.New(clientv3.Config{Endpoints: endpoints})
	if err != nil {
		panic(err)
	}

	s, err := concurrency.NewSession(cli, concurrency.WithTTL(5))
	if err != nil {
		fmt.Println("NewSession", "error", "err", err)
	}
	defer s.Close()

	// 创建一个新的etcd选举election
	e := concurrency.NewElection(s, "/resources/election")
	leaderCh := make(chan error)
	go m.elect(e, leaderCh)
	leaderChange := e.Observe(context.Background())
	select {
	case resp := <-leaderChange:
		m.logger.Info("watch leader change", zap.String("leader:", string(resp.Kvs[0].Value)))
	}
	workerNodeChange := m.WatchWorker()

	for {
		select {
		case err := <-leaderCh:
			if err != nil {
				m.logger.Error("leader elect failed", zap.Error(err))
				go m.elect(e, leaderCh)
			} else {
				m.logger.Info("master change to leader")
				m.leaderID = m.ID
				if !m.IsLeader() {
					m.BecomeLeader()
				}
			}
		case resp := <-leaderChange:
			if len(resp.Kvs) > 0 {
				m.logger.Info("watch leader change", zap.String("leader:", string(resp.Kvs[0].Value)))
			}
		case resp := <-workerNodeChange:
			m.logger.Info("watch worker change", zap.Any("worker:", resp))
			m.updateNodes()
		case <-time.After(20 * time.Second):
			rsp, err := e.Leader(context.Background())
			if err != nil {
				m.logger.Info("get Leader failed", zap.Error(err))
				if errors.Is(err, concurrency.ErrElectionNoLeader) {
					go m.elect(e, leaderCh)
				}
			}
			if rsp != nil && len(rsp.Kvs) > 0 {
				m.logger.Debug("get Leader", zap.String("value", string(rsp.Kvs[0].Value)))
				if m.IsLeader() && m.ID != string(rsp.Kvs[0].Value) {
					//当前已不再是leader
					atomic.StoreInt32(&m.ready, 0)
				}
			}
		}
	}
}

func (m *Master) elect(e *concurrency.Election, ch chan error) {
	// 堵塞直到选取成功
	err := e.Campaign(context.Background(), m.ID)
	ch <- err
}

func (m *Master) WatchWorker() chan *registry.Result {
	watch, err := m.registry.Watch(registry.WatchService(worker.ServiceName))
	if err != nil {
		panic(err)
	}
	ch := make(chan *registry.Result)
	go func() {
		for {
			res, err := watch.Next()
			if err != nil {
				m.logger.Info("watch worker service failed", zap.Error(err))
				continue
			}
			ch <- res
		}
	}()
	return ch

}
func (m *Master) BecomeLeader() {
	atomic.StoreInt32(&m.ready, 1)
}

func (m *Master) updateNodes() {
	services, err := m.registry.GetService(worker.ServiceName)
	if err != nil {
		m.logger.Error("get service", zap.Error(err))
	}

	nodes := make(map[string]*registry.Node)
	if len(services) > 0 {
		for _, spec := range services[0].Nodes {
			nodes[spec.Id] = spec
		}
	}

	added, deleted, changed := workNodeDiff(m.workNodes, nodes)
	m.logger.Sugar().Info("worker joined: ", added, ", leaved: ", deleted, ", changed: ", changed)

	m.workNodes = nodes

}

func workNodeDiff(old map[string]*registry.Node, new map[string]*registry.Node) ([]string, []string, []string) {
	added := make([]string, 0)
	deleted := make([]string, 0)
	changed := make([]string, 0)
	for k, v := range new {
		if ov, ok := old[k]; ok {
			if !reflect.DeepEqual(v, ov) {
				changed = append(changed, k)
			}
		} else {
			added = append(added, k)
		}
	}
	for k := range old {
		if _, ok := new[k]; !ok {
			deleted = append(deleted, k)
		}
	}
	return added, deleted, changed
}

// 获取本机网卡IP
func getLocalIP() (string, error) {
	var (
		addrs []net.Addr
		err   error
	)
	// 获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}
	// 取第一个非lo的网卡IP
	for _, addr := range addrs {
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("no local ip")
}
