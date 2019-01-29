package main

import (
	"github.com/go-redis/redis"

	"github.com/awesome-flow/flow/pkg/core"
	"github.com/awesome-flow/flow/pkg/devenv"
)

func main() {}

type RedisStreams struct {
	addr   string
	passwd string
	db     int
	*core.Connector
}

var _ core.Link = &RedisStreams{}

func New(name string, params core.Params, context *core.Context) (core.Link, error) {
	rs := &RedisStreams{
		"", "", 0,
		core.NewConnectorWithContext(context),
	}
	//TODO(olegs)

	rs.OnSetUp(rs.SetUp)
	rs.OnTearDown(rs.TearDown)

	return rs, nil
}

func (rs *RedisStreams) SetUp() error {
	ctx := rs.Connector.GetContext()
	var client *redis.Client
	client_i, ok := ctx.GetVal("redis-client")
	if ok {
		client = client_i.(*redis.Client)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     rs.addr,
			Password: rs.passwd,
			DB:       rs.db,
		})
		ctx.SetVal("redis-client", client)
	}

	//TODO(olegs)
	return nil
}

func (rs *RedisStreams) TearDown() error {
	//TODO(olegs)
	return nil
}

func (rs *RedisStreams) DevEnv(context *devenv.Context) ([]devenv.Fragment, error) {
	return []devenv.Fragment{
		devenv.DockerComposeFragment(`
  redis:
    image: redis:5.0.3-alpine3.8
    ports:
      - "6379:6379"
`),
	}, nil
}
