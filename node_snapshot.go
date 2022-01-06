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
	defer sink.Close()

	bytes, err := s.snapshot()
	if err != nil {
		return fmt.Errorf("s.snapshot(...): %w", err)
	}

	if _, err := sink.Write(bytes); err != nil {
		if err := sink.Cancel(); err != nil {
			return fmt.Errorf("sink.Cancel(...): %w", err)
		}

		return fmt.Errorf("sink.Write(...): %w", err)
	}

	return nil
}

func (s *fsmSnapshot) Release() {
}
