package main

import (
    "os"
    "bufio"
    "strings"
    "net"
    "strconv"
)

type RedisConfigServer struct {
    host string
    port uint16
}

type RedisConfig struct {
    servers []*RedisConfigServer
}

func NewConfig() *RedisConfig {
    return &RedisConfig {
        servers: make([]*RedisConfigServer, 0),
    }
}

func ConfigFromFile(path string) (*RedisConfig, error) {
    var err error
    cfg := NewConfig()

    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }

    reader := bufio.NewScanner(file)
    for reader.Scan() {
        serverLine := strings.TrimSpace(reader.Text())
        if len(serverLine) == 0 {
            continue
        }
        serverHost, serverPortString, err := net.SplitHostPort(serverLine)
        if err != nil {
            return nil, err
        }
        serverPort, err := strconv.ParseInt(serverPortString, 10, 16)
		if err != nil {
			return nil, err
		}
        server := &RedisConfigServer {
            host: serverHost,
            port: uint16(serverPort),
        }
        cfg.servers = append(cfg.servers, server)
    }

    return cfg, nil
}
