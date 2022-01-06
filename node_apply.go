package rafting

import (
	"context"
	"fmt"
	"time"

	pb "github.com/danielgatis/go-rafting/protobuf"
	"google.golang.org/grpc"
)

func applyOnLeader(n *Node, payload []byte, timeout time.Duration) (interface{}, error) {
	if n.raft.Leader() == "" {
		return nil, ErrUnknownleader
	}

	var opt grpc.DialOption = grpc.EmptyDialOption{}
	conn, err := grpc.Dial(string(n.raft.Leader()), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout), opt)
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
