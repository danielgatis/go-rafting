package rafting

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/danielgatis/go-discovery"
	"github.com/danielgatis/go-keyval"
	"github.com/hashicorp/raft"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type Node struct {
	id      string
	port    int
	dataDir string
	addr    string

	marshalFn   func(interface{}) ([]byte, error)
	unmarshalFn func([]byte, interface{}) error

	fsm           raft.FSM
	raft          *raft.Raft
	raftTransport *transport.Manager

	discovery               discovery.Discovery
	discoveryLookupInterval time.Duration

	logger     logrus.FieldLogger
	grpcServer *grpc.Server
}

func NewNode(ID string, fsm raft.FSM, port int, opts ...NodeOption) (*Node, error) {
	var (
		defaultDataDir                 = "./data"
		defaultLogger                  = logrus.StandardLogger()
		defaultDiscovery               = discovery.NewNullDiscovery()
		defaultDiscoveryLookupInterval = 2 * time.Second
		defaultMarshalFn               = json.Marshal
		defaultUnmarshalFn             = json.Unmarshal
	)

	n := &Node{
		id:                      ID,
		port:                    port,
		dataDir:                 defaultDataDir,
		logger:                  defaultLogger,
		discovery:               defaultDiscovery,
		discoveryLookupInterval: defaultDiscoveryLookupInterval,
		marshalFn:               defaultMarshalFn,
		unmarshalFn:             defaultUnmarshalFn,
		fsm:                     fsm,
	}

	for _, opt := range opts {
		opt(n)
	}

	if err := initRaft(n); err != nil {
		return nil, fmt.Errorf(`initRaft(...): %w`, err)
	}

	return n, nil
}

func (n *Node) Start(ctx context.Context) error {
	g := new(errgroup.Group)
	g.Go(func() error { return startRaft(ctx, n) })
	g.Go(func() error { return startDiscoveryLookup(ctx, n) })
	g.Go(func() error { return startDiscoveryRegister(ctx, n) })
	return g.Wait()
}

func (n *Node) Apply(cmd string, timeout time.Duration, args ...interface{}) (interface{}, error) {
	payload, err := n.marshalFn(&command{cmd, keyval.ToMap(args...)})
	if err != nil {
		return nil, fmt.Errorf(`n.marshalFn(...): %w`, err)
	}

	if err := n.raft.VerifyLeader().Error(); err != nil {
		return applyOnLeader(n, payload, timeout)
	}

	return apply(n, payload, timeout)
}
