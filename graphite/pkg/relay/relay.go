package relay

import (
	flow "github.com/whiteboxio/flow/pkg/core"
)

type Relay struct {
	Name string
	*flow.Connector
}

func NewRelay(name string, params flow.Params) (flow.Link, error) {
	return &Relay{
		name,
		flow.NewConnector(),
	}, nil
}
