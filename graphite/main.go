package main

import (
	"bytes"

	log "github.com/sirupsen/logrus"
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
	if ix := bytes.IndexByte(msg.Payload, ' '); ix != -1 {
		metricName := msg.Payload[:ix]
		msg.SetMeta("metric-name", metricName)
		log.Infof("Sending metric with name: [%s]", metricName)
		return mp.Send(msg)
	}
	return msg.AckInvalid()
}

func main() {}
