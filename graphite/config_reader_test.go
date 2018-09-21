package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestConfigReader_FromFile(t *testing.T) {

	tests := []struct {
		name        string
		config      string
		expectedCfg *GraphiteConfig
	}{
		{
			name: "single cluster, single route",
			config: `cluster 'test-cluster'
	jump_fnv1a_ch replication 1
		flow_dst_host:7220=000
		flow_dst_host:7221=001
		flow_dst_host:7222=002
		flow_dst_host:7223=003
		flow_dst_host:7224=004
	;
match .*
	send to
		test-cluster
	stop;`,
			expectedCfg: &GraphiteConfig{
				clusters: map[string]*GraphiteConfigCluster{
					"test-cluster": &GraphiteConfigCluster{
						name:       "test-cluster",
						ctype:      "jump_fnv1a_ch",
						replfactor: 1,
						servers: []*GraphiteConfigServer{
							&GraphiteConfigServer{
								host:  "flow_dst_host",
								port:  7220,
								index: 0,
							},
							&GraphiteConfigServer{
								host:  "flow_dst_host",
								port:  7221,
								index: 1,
							},
							&GraphiteConfigServer{
								host:  "flow_dst_host",
								port:  7222,
								index: 2,
							},
							&GraphiteConfigServer{
								host:  "flow_dst_host",
								port:  7223,
								index: 3,
							},
							&GraphiteConfigServer{
								host:  "flow_dst_host",
								port:  7224,
								index: 4,
							},
						},
					},
				},
				routes: []*GraphiteConfigRoute{
					&GraphiteConfigRoute{
						pattern:      regexp.MustCompile(".*"),
						destinations: []string{"test-cluster"},
						stop:         true,
						drop:         false,
					},
				},
			},
		},
		{
			name: "dual cluster, single route",
			config: `cluster test-cluster1
			jump_fnv1a_ch replication 3
				host_1_1:2001=001
				host_1_2:2002=002
				host_1_3:2003=003;
			
			cluster test-cluster2
			jump_fnv1a_ch replication 2
				host_2_2:2002=02
				host_2_3:2003=03
				host_2_1:2001=01;
			
			match metrics\..*
				send to
					test-cluster1
					test-cluster2;`,
			expectedCfg: &GraphiteConfig{
				clusters: map[string]*GraphiteConfigCluster{
					"test-cluster1": &GraphiteConfigCluster{
						name:       "test-cluster1",
						ctype:      "jump_fnv1a_ch",
						replfactor: 3,
						servers: []*GraphiteConfigServer{
							&GraphiteConfigServer{
								host:  "host_1_1",
								port:  2001,
								index: 1,
							},
							&GraphiteConfigServer{
								host:  "host_1_2",
								port:  2002,
								index: 2,
							},
							&GraphiteConfigServer{
								host:  "host_1_3",
								port:  2003,
								index: 3,
							},
						},
					},
					"test-cluster2": &GraphiteConfigCluster{
						name:       "test-cluster2",
						ctype:      "jump_fnv1a_ch",
						replfactor: 2,
						servers: []*GraphiteConfigServer{
							&GraphiteConfigServer{
								host:  "host_2_1",
								port:  2001,
								index: 1,
							},
							&GraphiteConfigServer{
								host:  "host_2_2",
								port:  2002,
								index: 2,
							},
							&GraphiteConfigServer{
								host:  "host_2_3",
								port:  2003,
								index: 3,
							},
						},
					},
				},
				routes: []*GraphiteConfigRoute{
					&GraphiteConfigRoute{
						pattern:      regexp.MustCompile("metrics\\..*"),
						destinations: []string{"test-cluster1", "test-cluster2"},
						stop:         false,
						drop:         false,
					},
				},
			},
		},
		{
			name: "dual cluster, dual route",
			config: `cluster test-cluster1
			jump_fnv1a_ch replication 3
				host_1_1:2001=001
				host_1_2:2002=002
				host_1_3:2003=003;
			
			cluster test-cluster2
			jump_fnv1a_ch replication 2
				host_2_2:2002=02
				host_2_3:2003=03
				host_2_1:2001=01;
			
			match metrics\..*
				send to
					test-cluster1
					test-cluster2;
			match .*
				drop;`,
			expectedCfg: &GraphiteConfig{
				clusters: map[string]*GraphiteConfigCluster{
					"test-cluster1": &GraphiteConfigCluster{
						name:       "test-cluster1",
						ctype:      "jump_fnv1a_ch",
						replfactor: 3,
						servers: []*GraphiteConfigServer{
							&GraphiteConfigServer{
								host:  "host_1_1",
								port:  2001,
								index: 1,
							},
							&GraphiteConfigServer{
								host:  "host_1_2",
								port:  2002,
								index: 2,
							},
							&GraphiteConfigServer{
								host:  "host_1_3",
								port:  2003,
								index: 3,
							},
						},
					},
					"test-cluster2": &GraphiteConfigCluster{
						name:       "test-cluster2",
						ctype:      "jump_fnv1a_ch",
						replfactor: 2,
						servers: []*GraphiteConfigServer{
							&GraphiteConfigServer{
								host:  "host_2_1",
								port:  2001,
								index: 1,
							},
							&GraphiteConfigServer{
								host:  "host_2_2",
								port:  2002,
								index: 2,
							},
							&GraphiteConfigServer{
								host:  "host_2_3",
								port:  2003,
								index: 3,
							},
						},
					},
				},
				routes: []*GraphiteConfigRoute{
					&GraphiteConfigRoute{
						pattern:      regexp.MustCompile("metrics\\..*"),
						destinations: []string{"test-cluster1", "test-cluster2"},
						stop:         false,
						drop:         false,
					},
					&GraphiteConfigRoute{
						pattern: regexp.MustCompile(".*"),
						stop:    false,
						drop:    true,
					},
				},
			},
		},
	}

	t.Parallel()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tmpFile, err := ioutil.TempFile("/tmp", "flow-graphite-test-config")
			if err != nil {
				t.Fatalf("Failed to create a tmp file: %s", err)
			}
			defer os.Remove(tmpFile.Name())

			if err := ioutil.WriteFile(tmpFile.Name(), []byte(testCase.config), 0644); err != nil {
				t.Fatalf("Failed to write the data to tmp file: %s", err)
			}

			cfg, err := ConfigFromFile(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to read the config: %s", err)
			}

			if !reflect.DeepEqual(cfg.clusters, testCase.expectedCfg.clusters) {
				t.Errorf("Diverging config cluster values: want: %+v, got: %+v",
					cfg.clusters, testCase.expectedCfg.clusters)
			}

			if !reflect.DeepEqual(cfg.routes, testCase.expectedCfg.routes) {
				t.Errorf("Diverging config route values: want: %+v, got: %+v",
					cfg.routes, testCase.expectedCfg.routes)
				for i, route := range cfg.routes {
					if i >= len(testCase.expectedCfg.routes) {
						t.Errorf("Expected config is missing index %d", i)
					}
					if !reflect.DeepEqual(route, testCase.expectedCfg.routes[i]) {
						t.Errorf("Diverging route value at index %d: got %+v, want: %+v",
							i, route, testCase.expectedCfg.routes[i])
					}
				}
			}

			if !reflect.DeepEqual(cfg, testCase.expectedCfg) {
				t.Fatalf("Unexpected config value: %+v, want: %+v", cfg, testCase.expectedCfg)
			}
		})
	}

}
