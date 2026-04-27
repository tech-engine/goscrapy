package executor

import "errors"

var (
	ErrAdapterRequired = errors.New("executor: adapter is required in config")
)
