package main

import (
	"bytes"
	"fmt"

	"github.com/awesome-flow/flow/pkg/metrics"

	"github.com/go-redis/redis"

	"github.com/awesome-flow/flow/pkg/core"
	"github.com/awesome-flow/flow/pkg/devenv"
)

func main() {}

const MetricsBasePath = "plugins.sink.redis-streams"

type RedisStreams struct {
	address  string
	password string
	db       int
	client   *redis.Client
	*core.Connector
}

var _ core.Link = &RedisStreams{}

func New(name string, params core.Params, context *core.Context) (core.Link, error) {

	var address string
	if v, ok := params["address"]; ok {
		address = v.(string)
	} else {
		return nil, fmt.Errorf("redis-streams is missing a mandatory argument addres")
	}

	var password string
	if v, ok := params["password"]; ok {
		password = v.(string)
	} else {
		password = ""
	}

	var db int
	if v, ok := params["db"]; ok {
		db = v.(int)
	} else {
		db = 0
	}

	rs := &RedisStreams{
		address, password, db,
		nil, core.NewConnectorWithContext(context),
	}

	rs.OnSetUp(rs.SetUp)
	rs.OnTearDown(rs.TearDown)

	return rs, nil
}

func (rs *RedisStreams) SetUp() error {
	ctx := rs.Connector.GetContext()
	var client *redis.Client
	iclient, ok := ctx.GetVal("redis-client")
	if ok {
		client = iclient.(*redis.Client)
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     rs.address,
			Password: rs.password,
			DB:       rs.db,
		})
		ctx.SetVal("redis-client", client)
	}
	rs.client = client

	return nil
}

func (rs *RedisStreams) TearDown() error {
	if rs.client != nil {
		return rs.client.Close()
	}
	return nil
}

func (rs *RedisStreams) Recv(msg *core.Message) error {
	metrics.GetCounter(MetricsBasePath + ".received").Inc(1)
	if rs.client == nil {
		metrics.GetCounter(MetricsBasePath + ".conn_failed").Inc(1)
		return msg.AckFailed()
	}
	xargs, err := ParseXAddMsg(msg)
	if err != nil {
		metrics.GetCounter(MetricsBasePath + ".bad_format").Inc(1)
		return msg.AckFailed()
	}
	if err := rs.client.XAdd(xargs).Err(); err != nil {
		metrics.GetCounter(MetricsBasePath + ".send_failed").Inc(1)
		return msg.AckFailed()
	}
	metrics.GetCounter(MetricsBasePath + ".sent").Inc(1)
	return msg.AckDone()
}

var splitter = []byte(" ")

// ParseXAddMsg takes a pointer to a message and parses it into
// Redis XAdd arguments according to the format:
//   mystream * sensor-id 1234 temperature 19.8
// The mesage payload should have at least 2 words: the stream
// name and the ID (optional, * if empty).
// Keys and values are converted into string tuples.
// Keys are expected to be unique, last key wins.
func ParseXAddMsg(msg *core.Message) (*redis.XAddArgs, error) {
	chunks := bytes.Split(msg.Payload(), splitter)
	if len(chunks) < 2 || (len(chunks)%2 > 0) {
		return nil, fmt.Errorf("Malformed XAdd message: %q", msg.Payload())
	}
	xargs := &redis.XAddArgs{
		Stream: string(chunks[0]),
		Values: make(map[string]interface{}),
	}

	if string(chunks[1]) != "*" {
		xargs.ID = string(chunks[1])
	}

	chunks = chunks[2:]
	for len(chunks) > 0 {
		key, value := string(chunks[0]), string(chunks[1])
		xargs.Values[key] = value
		chunks = chunks[2:]
	}

	return xargs, nil
}

func (rs *RedisStreams) ConnectTo(core.Link) error {
	panic("Redis-streams sink is not supposed to be connnected")
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
