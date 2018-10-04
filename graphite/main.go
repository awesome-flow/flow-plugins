package main

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/whiteboxio/flow/pkg/core"
	replicator "github.com/whiteboxio/flow/pkg/link/replicator"
	tcp_sink "github.com/whiteboxio/flow/pkg/sink/tcp"
)

const (
	MsgSendTimeout = 100 * time.Millisecond
)

type GraphiteLink struct {
	name     string
	config   *GraphiteConfig
	clusters map[string]core.Link
	*core.Connector
}

func New(name string, params core.Params) (core.Link, error) {
	link, err := bootstrap(name, params)
	return link, err
}

func (gl *GraphiteLink) Recv(msg *core.Message) error {
	var metricName string
	if ix := bytes.IndexByte(msg.Payload, ' '); ix != -1 {
		msg.SetMeta("metric-name", msg.Payload[:ix])
		metricName = string(msg.Payload[:ix])
	} else {
		return msg.AckUnroutable()
	}
	var succCnt, totalCnt, failCnt int32 = 0, 0, 0
	wg := &sync.WaitGroup{}
Routes:
	for _, route := range gl.config.routes {
		if route.pattern.MatchString(metricName) {
		Dst:
			for _, dst := range route.destinations {
				endpoint, ok := gl.clusters[dst]
				if !ok {
					continue Dst
				}
				msgCp := core.CpMessage(msg)
				wg.Add(1)
				totalCnt++
				go func() {
					if err := endpoint.Recv(msgCp); err != nil {
						atomic.AddInt32(&failCnt, 1)
					} else {
						atomic.AddInt32(&succCnt, 1)
					}
					wg.Done()
				}()
			}
			if route.stop {
				break Routes
			}
		}
	}
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
		close(done)
	}()
	select {
	case <-done:
		if failCnt != 0 {
			if failCnt == totalCnt {
				return msg.AckFailed()
			} else {
				return msg.AckPartialSend()
			}
		} else {
			return msg.AckDone()
		}
	case <-time.After(MsgSendTimeout):
		return msg.AckTimedOut()
	}
}

func bootstrap(name string, params core.Params) (core.Link, error) {
	configPath, ok := params["config"]
	if !ok {
		return nil, fmt.Errorf("Missing graphite config path")
	}
	config, err := ConfigFromFile(configPath.(string))
	if err != nil {
		return nil, err
	}
	clusters := make(map[string]core.Link)
	for name, cfg := range config.clusters {
		cluster, err := buildCluster(cfg)
		if err != nil {
			return nil, err
		}
		clusters[name] = cluster
	}
	graphite := &GraphiteLink{
		name,
		config,
		clusters,
		core.NewConnector(),
	}

	return graphite, nil
}

func buildCluster(config *GraphiteConfigCluster) (core.Link, error) {
	repl, err := replicator.New(config.name, core.Params{
		"hash_key":  "metric-name",
		"hash_algo": config.ctype,
		"replicas":  config.replfactor,
	})
	if err != nil {
		return nil, err
	}
	endpoints := make([]core.Link, len(config.servers))
	for ix, serverCfg := range config.servers {
		endpoint, err := tcp_sink.New(
			fmt.Sprintf("graphite_endpoint_%s_%d", serverCfg.host, serverCfg.port),
			core.Params{
				"bind_addr": fmt.Sprintf("%s:%d", serverCfg.host, serverCfg.port),
			},
		)
		if err != nil {
			return nil, err
		}
		endpoints[ix] = endpoint
	}
	if err := repl.LinkTo(endpoints); err != nil {
		return nil, err
	}

	return repl, nil
}

func main() {}
