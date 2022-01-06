package rafting

import (
	"fmt"
)

var (
	ErrUnknownleader  = fmt.Errorf("Unknown leader")
	ErrInvalidCommand = fmt.Errorf("Invalid cmd")
)
