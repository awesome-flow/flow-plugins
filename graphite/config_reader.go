package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type GraphiteConfigAggregator struct {
	//TODO
}

type GraphiteConfigCluster struct {
	name       string
	ctype      string
	replfactor uint
	servers    []*GraphiteConfigServer
}

type GraphiteConfigRoute struct {
	pattern      *regexp.Regexp
	destinations []string
	stop         bool
	drop         bool
}

type GraphiteConfigServer struct {
	host  string
	port  uint16
	index uint32
}

type GraphiteConfig struct {
	//aggregators []*GraphiteConfigAggregator
	clusters map[string]*GraphiteConfigCluster
	routes   []*GraphiteConfigRoute
}

func NewConfig() *GraphiteConfig {
	return &GraphiteConfig{
		clusters: make(map[string]*GraphiteConfigCluster),
		routes:   make([]*GraphiteConfigRoute, 0),
	}
}

func ConfigFromFile(path string) (*GraphiteConfig, error) {
	var err error
	cfg := NewConfig()

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewScanner(file)

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if len(line) == 0 {
			continue
		}
		if match, _ := regexp.MatchString("^cluster", line); match {
			cluster := &GraphiteConfigCluster{
				name: strings.Trim(line[7:], " '\""),
			}
			replRegex, err := regexp.Compile("([\\w\\d_]+)\\s+replication\\s+(\\d+)")
			if err != nil {
				return nil, err
			}
			if !reader.Scan() {
				return nil, fmt.Errorf("Expected to get a cluster definition "+
					"for cluster %q, got none", cluster.name)
			}
			replLine := reader.Text()
			replMatch := replRegex.FindStringSubmatch(replLine)
			if len(replMatch) == 0 {
				return nil, fmt.Errorf("Unexpected replication definition found: "+
					"%q, can not parse it", replLine)
			}
			hashAlg, replFactorStr := replMatch[1], replMatch[2]
			replFactor, err := strconv.ParseUint(replFactorStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse replication factor: %q",
					replFactorStr)
			}
			cluster.replfactor = uint(replFactor)
			cluster.ctype = hashAlg
			clusterServerRegex := regexp.MustCompile("^(\\w+):(\\d+)=(\\d+)$")
		Cluster:
			for reader.Scan() {
				clusterLine := strings.TrimSpace(reader.Text())
				shouldBreak := false
				if match, _ := regexp.MatchString(";$", clusterLine); match {
					shouldBreak = true
					clusterLine = strings.Trim(clusterLine, ";")
				}
				if len(clusterLine) != 0 {
					match := clusterServerRegex.FindStringSubmatch(clusterLine)
					if len(match) == 0 {
						return nil, fmt.Errorf("Failed to parse cluster server"+
							" config: %q", clusterLine)
					}
					serverPort, err := strconv.ParseUint(match[2], 10, 16)
					if err != nil {
						return nil, err
					}
					serverIndex, err := strconv.ParseUint(match[3], 10, 32)
					if err != nil {
						return nil, err
					}
					server := &GraphiteConfigServer{
						host:  match[1],
						port:  uint16(serverPort),
						index: uint32(serverIndex),
					}
					cluster.servers = append(cluster.servers, server)
				}
				if shouldBreak {
					sort.Slice(cluster.servers, func(i, j int) bool {
						return cluster.servers[i].index < cluster.servers[j].index
					})
					break Cluster
				}
			}
			cfg.clusters[cluster.name] = cluster
		} else if match, _ := regexp.MatchString("^match", line); match {
			matchRule := strings.TrimSpace(line[5:])
			fmt.Printf("Route definition started: %q\n", matchRule)
			matchRegex, err := regexp.Compile(matchRule)
			if err != nil {
				return nil, err
			}
			configRoute := &GraphiteConfigRoute{
				pattern: matchRegex,
			}
			shouldBreak := false
			sendToStarted := false
		Match:
			for reader.Scan() {
				line := strings.TrimSpace(reader.Text())
				if match, _ := regexp.MatchString(";$", line); match {
					shouldBreak = true
					line = strings.Trim(line, ";")
				}
				if len(line) > 0 {
					fmt.Printf("Match line: %q\n", line)
					if line == "stop" {
						configRoute.stop = true
					} else if line == "drop" {
						configRoute.drop = true
					} else if line == "send to" {
						sendToStarted = true
					} else if sendToStarted {
						configRoute.destinations = append(configRoute.destinations, line)
					} else {
						return nil, fmt.Errorf("Unexpected line: %q", line)
					}
				}
				if shouldBreak {
					break Match
				}
			}
			cfg.routes = append(cfg.routes, configRoute)
		}
	}
	return cfg, nil
}
