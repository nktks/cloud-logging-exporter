package agent

import "context"

type Agent interface {
	Run(context.Context) error
}
