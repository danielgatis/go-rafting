package rafting

import (
	context "context"
	"fmt"

	pb "github.com/danielgatis/go-rafting/protobuf"
)

type raftingServiceServer struct {
	pb.UnimplementedRaftingServiceServer
	node *Node
}

func newRaftingServerImpl(node *Node) *raftingServiceServer {
	return &raftingServiceServer{
		node: node,
	}
}

func (s *raftingServiceServer) Apply(ctx context.Context, request *pb.ApplyRequest) (*pb.ApplyResponse, error) {
	result := s.node.raft.Apply(request.Payload, 0)
	if result.Error() != nil {
		return nil, fmt.Errorf("raft.Apply(...): %w", result.Error())
	}

	payload, err := s.node.marshalFn(result.Response())
	if err != nil {
		return nil, fmt.Errorf("node.marshalFn(...): %w", err)
	}

	return &pb.ApplyResponse{Payload: payload}, nil
}

func (s *raftingServiceServer) GetDetails(context.Context, *pb.GetDetailsRequest) (*pb.GetDetailsResponse, error) {
	return &pb.GetDetailsResponse{
		Id:   s.node.id,
		Port: int32(s.node.port),
	}, nil
}
