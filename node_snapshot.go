package rafting

import (
	"fmt"

	"github.com/hashicorp/raft"
)

var _ raft.FSMSnapshot = (*fsmSnapshot)(nil)

type snapshotFunc = func() ([]byte, error)

type fsmSnapshot struct {
	snapshot snapshotFunc
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		bytes, err := s.snapshot()
		if err != nil {
			return fmt.Errorf("state.takeSnapshot(...): %w", err)
		}

		if _, err := sink.Write(bytes); err != nil {
			return fmt.Errorf("sink.Write(...): %w", err)
		}

		if err := sink.Close(); err != nil {
			return fmt.Errorf("sink.Close(...): %w", err)
		}

		return nil
	}()

	if err != nil {
		if err := sink.Cancel(); err != nil {
			return fmt.Errorf("sink.Cancel(...): %w", err)
		}

		return err
	}

	return nil
}

func (s *fsmSnapshot) Release() {
}
