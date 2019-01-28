package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/whiteboxio/flow/pkg/core"
)

type counterRecv struct {
	cnt uint64
	*core.Connector
}

func newCounterRecv() *counterRecv {
	return &counterRecv{0, core.NewConnector()}
}

func (cr *counterRecv) Recv(msg *core.Message) error {
	cr.cnt++
	return nil
}

type testRecv struct {
	*GraphiteLink
}

func newTestRecv() *testRecv {
	cfgPath := "/home/olegs/workspace/golang/src/github.com/whiteboxio/flow-plugins/graphite/configs/graphite-relay.conf"
	recv, err := New(
		"test graphite receiver",
		core.Params{"config": cfgPath},
		core.NewContext())
	if err != nil {
		panic(err.Error())
	}
	return &testRecv{recv.(*GraphiteLink)}
}

func (tr *testRecv) buildCluster(config *GraphiteConfigCluster) (core.Link, error) {
	return newCounterRecv(), nil
}

func BenchmarkRecv(b *testing.B) {
	r := newTestRecv()
	for i := 0; i < b.N; i++ {
		payload := fmt.Sprintf("foo.bar.baz 42 %d", time.Now().Unix())
		msg := core.NewMessage([]byte(payload))
		if err := r.Recv(msg); err != nil {
			panic(err.Error())
		}
	}
}
