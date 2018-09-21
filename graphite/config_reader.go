package main

import (
	"net"
)

type GraphiteConfigAggregator struct {
	//TODO
}

type GraphiteConfigCluster struct {
	name       string
	ctype      string
	replfactor uint
	servers    []*GraphiteConfigServer
	next       *GraphiteConfigCluster
}

type GraphiteConfigRoute struct {
	pattern      string
	destinations []*GraphiteConfigCluster
	stop         bool
	next         *GraphiteConfigRoute
}

type GraphiteConfigServer struct {
	ip   net.IP
	port uint8
}

type GraphiteConfig struct {
	aggregators []GraphiteConfigAggregator
	clusters    []GraphiteConfigCluster
	routes      []GraphiteConfigRoute
	servers     []GraphiteConfigServer
}

func New() *GraphiteConfig {
	return &GraphiteConfig{
		aggregators: make([]GraphiteConfigAggregator, 0),
		clusters:    make([]GraphiteConfigCluster, 0),
		routes:      make([]GraphiteConfigRoute, 0),
		servers:     make([]GraphiteConfigServer, 0),
	}
}

func ReadFile(path string) (*GraphiteConfig, error) {
	//TODO
	cfg := New()
	return cfg, nil
}
