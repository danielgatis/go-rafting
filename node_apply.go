package rafting

import (
	"context"
	"fmt"
	"time"

	pb "github.com/danielgatis/go-rafting/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func applyOnLeader(n *Node, payload []byte, timeout time.Duration) (interface{}, error) {
	if n.raft.Leader() == "" {
		return nil, ErrUnknownleader
	}

	var opt grpc.DialOption = grpc.EmptyDialOption{}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, string(n.raft.Leader()), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), opt)
	if err != nil {
		return nil, fmt.Errorf(`grpc.Dial(...): %w`, err)
	}

	defer conn.Close()
	client := pb.NewRaftingServiceClient(conn)

	response, err := client.Apply(context.Background(), &pb.ApplyRequest{Payload: payload})
	if err != nil {
		return nil, fmt.Errorf(`client.Apply(...): %w`, err)
	}

	var result interface{}
	err = n.unmarshalFn(response.Payload, &result)
	if err != nil {
		return nil, fmt.Errorf("n.unmarshalFn(...): %w", err)
	}

	return result, nil
}

func apply(n *Node, payload []byte, timeout time.Duration) (interface{}, error) {
	result := n.raft.Apply(payload, timeout)
	if result.Error() != nil {
		return nil, result.Error()
	}

	switch result.Response().(type) {
	case error:
		return nil, result.Response().(error)
	default:
		return result.Response(), nil
	}
}
