package rafting

import (
	"bytes"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

var _ raft.FSM = (*fsm)(nil)

type fsm struct {
	state *State
}

func (f *fsm) Apply(log *raft.Log) interface{} {
	var cmd command

	if err := f.state.unmarshalFn(log.Data, &cmd); err != nil {
		return fmt.Errorf(`state.unmarshalFn(...): %w`, err)
	}

	if handler, ok := f.state.commands[cmd.Name]; ok {
		result, err := handler(f.state.data, cmd.Args)
		if err != nil {
			return err
		}

		return result
	}

	return ErrInvalidCommand
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{func() ([]byte, error) {
		b, err := f.state.marshalFn(f.state.data)
		if err != nil {
			return nil, fmt.Errorf(`state.marshalFn(...): %w`, err)
		}

		return b, nil
	}}, nil
}

func (f *fsm) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(rc); err != nil {
		return fmt.Errorf(`buf.ReadFrom(...): %w`, err)
	}

	if err := f.state.unmarshalFn(buf.Bytes(), &f.state.data); err != nil {
		return fmt.Errorf(`state.unmarshalFn(...): %w`, err)
	}

	return nil
}
