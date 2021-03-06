package rafting

import (
	"context"
	"fmt"
	"time"

	pb "github.com/danielgatis/go-rafting/protobuf"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type peer struct {
	id   string
	addr string
	port int32
}

func getPeers(addrs []string) []peer {
	peers := make([]peer, 0)

	for _, addr := range addrs {
		p, err := getPeer(addr)
		if err != nil {
			continue
		}

		peers = append(peers, peer{
			id:   p.Id,
			addr: addr,
			port: p.Port,
		})
	}

	return peers
}

func getPeer(addr string) (*pb.GetDetailsResponse, error) {
	var opt grpc.DialOption = grpc.EmptyDialOption{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), opt)
	if err != nil {
		return nil, fmt.Errorf(`grpc.Dial(...): %w`, err)
	}

	defer conn.Close()
	client := pb.NewRaftingServiceClient(conn)

	response, err := client.GetDetails(context.Background(), &pb.GetDetailsRequest{})
	if err != nil {
		return nil, fmt.Errorf(`client.GetDetails(...): %w`, err)
	}

	return response, nil
}

func remPeer(n *Node, details []peer) {
	for _, server := range n.raft.GetConfiguration().Configuration().Servers {
		found := false
		for _, detail := range details {
			if string(server.Address) == detail.addr || string(server.ID) == detail.id {
				found = true
				break
			}
		}

		if !found {
			n.raft.RemoveServer(server.ID, 0, 0)
		}
	}
}

func addPeer(n *Node, details []peer) {
	for _, detail := range details {
		found := false
		for _, server := range n.raft.GetConfiguration().Configuration().Servers {
			if string(server.Address) == detail.addr || string(server.ID) == detail.id {
				found = true
				break
			}
		}

		if !found {
			n.raft.AddVoter(raft.ServerID(detail.id), raft.ServerAddress(detail.addr), 0, 0)
		}
	}
}
