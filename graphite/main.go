package main

import (
	"bytes"

	core "github.com/whiteboxio/flow/pkg/core"
)

type MsgParser struct {
	Name string
	*core.Connector
}

func NewMsgParser(name string, params core.Params) (core.Link, error) {
	return &MsgParser{name, core.NewConnector()}, nil
}

func (mp *MsgParser) Recv(msg *core.Message) error {
	//fmt.Printf("Graphite plugin received a new message: {%s}\n", msg.Payload)
	if ix := bytes.IndexByte(msg.Payload, ' '); ix != -1 {
		metricName := msg.Payload[:ix]
		msg.SetMeta("metric-name", metricName)
		//log.Infof("Sending metric with name: [%s]", metricName)
		return mp.Send(msg)
	}
	return msg.AckInvalid()
}

func main() {}
