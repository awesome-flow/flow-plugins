package main

import (
    "fmt"
    "github.com/awesome-flow/flow/pkg/core"
    "github.com/go-redis/redis"
    "strconv"
    "bytes"
)

type RedisLink struct {
    name string
    config *RedisConfig
    endpoints []*redis.Client
    *core.Connector
}

func New(name string, params core.Params, context *core.Context) (core.Link, error) {
    link, err := bootstrap(name, params, context)
    return link, err
}

func bootstrap(name string, params core.Params, context *core.Context) (core.Link, error) {
    configPath, ok := params["config"]
    if !ok {
        return nil, fmt.Errorf("Missing redis config path")
    }
    config, err := ConfigFromFile(configPath.(string))
    if err != nil {
        return nil, err
    }
    redisEndPoints, err := buildRedisConn(config)
    if err != nil {
        return nil, err
    }
    redisLink := &RedisLink {
        name,
        config,
        redisEndPoints,
        core.NewConnectorWithContext(context),
    }
    return redisLink, nil
}

// Format of message payload (stream_name   id   key   value)
func (rl *RedisLink) Recv(msg *core.Message) error {
    messageParts := bytes.Split(msg.Payload, []byte(" "))
    if len(messageParts) != 4 {
        return fmt.Errorf("Invalid message payload. It should be of length 4 in the format (stream_name id key value)")
    }
    redisClient := rl.endpoints[0] // For now, let's consider the first one. Later, we will migrate this to a cluster

    err := sendMessageToRedisStream(redisClient, messageParts)
    if err != nil {
        return msg.AckFailed()
    }

    return msg.AckDone()
}


func sendMessageToRedisStream(rClient *redis.Client, messageParts[][]byte) error {
    streamTopic, streamId, key, value := messageParts[0], messageParts[1], messageParts[2], messageParts[3]

    streamValues := map[string]interface{}{string(key) : string(value)}

    redisStreamData := &redis.XAddArgs {
        Stream: string(streamTopic),
        ID: string(streamId),
        Values: streamValues,
    }
    err := rClient.XAdd(redisStreamData).Err()
    if err != nil {
        return err
    }

    return nil
}

func buildRedisConn(config *RedisConfig) ([]*redis.Client, error) {
    endpoints := make([]*redis.Client, len(config.servers))

	for ix, serverCfg := range config.servers {
        serverHost, serverPort := serverCfg.host, serverCfg.port
        redisEndPoint := redis.NewClient(&redis.Options {
            Addr: serverHost + ":" + strconv.Itoa(int(serverPort)),
            Password: "",
            DB: 0,
        })
		endpoints[ix] = redisEndPoint
	}

    return endpoints, nil
}

func main() {}
