package pipelines

import "context"

type OpenHook func(context.Context) error

type CloseHook func()
