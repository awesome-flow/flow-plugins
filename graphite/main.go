package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/awesome-flow/flow/pkg/core"
	demux "github.com/awesome-flow/flow/pkg/link/demux"
	replicator "github.com/awesome-flow/flow/pkg/link/replicator"
	tcp_sink "github.com/awesome-flow/flow/pkg/sink/tcp"
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

func New(name string, params core.Params, context *core.Context) (core.Link, error) {
	link, err := bootstrap(name, params, context)
	return link, err
}

func (gl *GraphiteLink) Recv(msg *core.Message) error {
	var metricName string
	payload := msg.Payload()
	if ix := bytes.IndexByte(payload, ' '); ix != -1 {
		msg.SetMeta("metric-name", payload[:ix])
		metricName = string(payload[:ix])
	} else {
		return msg.AckUnroutable()
	}
	dests := make(map[string]bool)
Routes:
	for _, route := range gl.config.routes {
		if route.pattern.MatchString(metricName) {
			for _, dest := range route.destinations {
				dests[dest] = true
			}
			if route.drop {
				continue Routes
			}
			if route.stop {
				break Routes
			}
		}
	}
	links := make([]core.Link, len(dests))
	ix := 0
	for dest := range dests {
		links[ix] = gl.clusters[dest]
		ix++
	}

	return demux.Demultiplex(msg, 1<<uint(ix)-1, links[:ix], MsgSendTimeout)
}

func bootstrap(name string, params core.Params, context *core.Context) (core.Link, error) {
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
		core.NewConnectorWithContext(context),
	}

	return graphite, nil
}

func buildCluster(config *GraphiteConfigCluster) (core.Link, error) {
	ctx := core.NewContext()
	repl, err := replicator.New(
		config.name,
		core.Params{
			"hash_key":  "metric-name",
			"hash_algo": config.ctype,
			"replicas":  config.replfactor,
		},
		ctx,
	)
	if err != nil {
		return nil, err
	}
	endpoints := make([]core.Link, len(config.servers))
	for ix, serverCfg := range config.servers {
		tcpCtx := core.NewContext()
		endpoint, err := tcp_sink.New(
			fmt.Sprintf("graphite_endpoint_%s_%d", serverCfg.host, serverCfg.port),
			core.Params{
				"bind_addr": fmt.Sprintf("%s:%d", serverCfg.host, serverCfg.port),
			},
			tcpCtx,
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
