package rafting

import (
	"encoding/json"

	"github.com/hashicorp/raft"
)

type stateCommandFunc = func(data interface{}, args map[string]interface{}) (interface{}, error)

type State struct {
	data     interface{}
	commands map[string]stateCommandFunc

	marshalFn   func(interface{}) ([]byte, error)
	unmarshalFn func([]byte, interface{}) error
}

func NewState(data interface{}) *State {
	return NewStateWithSerializer(data, json.Marshal, json.Unmarshal)
}

func NewStateWithSerializer(data interface{}, marshalFn func(interface{}) ([]byte, error), unmarshalFn func([]byte, interface{}) error) *State {
	return &State{
		data:     data,
		commands: make(map[string]stateCommandFunc),

		marshalFn:   marshalFn,
		unmarshalFn: unmarshalFn,
	}
}

func (s *State) Command(cmd string, handler stateCommandFunc) {
	s.commands[cmd] = handler
}

func (s *State) FSM() raft.FSM {
	return &fsm{s}
}
