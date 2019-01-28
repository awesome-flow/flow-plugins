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
	rsl := &RedisStreams{
		"", "", 0,
		core.NewConnectorWithContext(context),
	}
	//TODO(olegs)

	rsl.OnSetUp(rsl.SetUp)
	rsl.OnTearDown(rsl.TearDown)

	return rsl, nil
}

func (rsl *RedisStreams) SetUp() error {
	ctx := rsl.Connector.GetContext()
	var client *redis.Client
	client_i, ok := ctx.GetVal("redis-client")
	if ok {
		client = client_i.(*redis.Client)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     rsl.addr,
			Password: rsl.passwd,
			DB:       rsl.db,
		})
		ctx.SetVal("redis-client", client)
	}

	//TODO(olegs)
	return nil
}

func (rsl *RedisStreams) TearDown() error {
	//TODO(olegs)
	return nil
}

func (rsl *RedisStreams) DevEnv(context *devenv.Context) ([]devenv.Fragment, error) {
	return []devenv.Fragment{
		devenv.DockerComposeFragment(`
  redis:
    image: redis:5.0.3-alpine3.8
    ports:
      - "6379:6379"
`),
	}, nil
}
