package rafting

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	transport "github.com/Jille/raft-grpc-transport"
	adapters "github.com/danielgatis/go-logrus-adapters"
	pb "github.com/danielgatis/go-rafting/protobuf"
	"github.com/hashicorp/raft"
	store "github.com/hashicorp/raft-boltdb/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initRaft(n *Node) error {
	if err := os.MkdirAll(filepath.Join(n.dataDir, n.id), 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("os.MkdirAll(...): %w", err)
	}

	logStore, err := store.NewBoltStore(filepath.Join(n.dataDir, n.id, "log-store.db"))
	if err != nil {
		return fmt.Errorf("store.NewBoltStore(...): %w", err)
	}

	stableStore, err := store.NewBoltStore(filepath.Join(n.dataDir, n.id, "stable-store.db"))
	if err != nil {
		return fmt.Errorf("store.NewBoltStore(...): %w", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(filepath.Join(n.dataDir, n.id), n.snapshotRetain, os.Stderr)
	if err != nil {
		return fmt.Errorf("raft.NewFileSnapshotStore(...): %w", err)
	}

	conf := raft.DefaultConfig()
	conf.LocalID = raft.ServerID(n.id)
	conf.Logger = adapters.NewHCLogAdapter(n.logger, "raft")

	n.addr = fmt.Sprintf("%s:%d", "0.0.0.0", n.port)
	n.grpcServer = grpc.NewServer()
	n.raftTransport = transport.New(raft.ServerAddress(n.addr), []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
	n.raftTransport.Register(n.grpcServer)
	pb.RegisterRaftingServiceServer(n.grpcServer, newRaftingServerImpl(n))

	n.raft, err = raft.NewRaft(conf, n.fsm, logStore, stableStore, snapshotStore, n.raftTransport.Transport())
	if err != nil {
		return fmt.Errorf(`raft.NewRaft(...): %w`, err)
	}

	return nil
}

func startRaft(ctx context.Context, n *Node) error {
	go func() {
		<-ctx.Done()
		n.grpcServer.GracefulStop()

		if err := n.raft.Snapshot().Error(); err != nil {
			if err != raft.ErrNothingNewToSnapshot {
				n.logger.Error(err)
			}
		}

		if err := n.raft.Shutdown().Error(); err != nil {
			n.logger.Error(err)
		}
	}()

	if f := n.raft.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(n.id),
				Address: n.raftTransport.Transport().LocalAddr(),
			},
		},
	}); f.Error() != nil && f.Error() != raft.ErrCantBootstrap {
		return fmt.Errorf(`raft.BootstrapCluster(...): %w`, f.Error())
	}

	listen, err := net.Listen("tcp", n.addr)
	if err != nil {
		return fmt.Errorf(`net.Listen(...): %w`, err)
	}

	if err := n.grpcServer.Serve(listen); err != nil {
		return fmt.Errorf(`grpcServer.Serve(...): %w`, err)
	}

	return nil
}
