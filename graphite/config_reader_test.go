package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func TestConfigReader_FromFile(t *testing.T) {

	configData := []byte(`
cluster 'test-cluster'
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
	stop;

`)

	tmpFile, err := ioutil.TempFile("/tmp", "flow-graphite-test-config")
	if err != nil {
		t.Fatalf("Failed to create a tmp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := ioutil.WriteFile(tmpFile.Name(), configData, 0644); err != nil {
		t.Fatalf("Failed to write the data to tmp file: %s", err)
	}

	cfg, err := FromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read the config: %s", err)
	}

	expectedCfg := &GraphiteConfig{
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
	}
	if !reflect.DeepEqual(cfg.clusters, expectedCfg.clusters) {
		t.Fatalf("Mismatch in config clsuters: got: %+v, want: %+v",
			cfg.clusters, expectedCfg.clusters)
	}
	if !reflect.DeepEqual(cfg.routes, expectedCfg.routes) {
		t.Fatalf("Mismatch in config routes: got: %+v, want: %+v",
			cfg.routes, expectedCfg.routes)
	}
	//if !reflect.DeepEqual(cfg, expectedCfg) {
	//	t.Fatalf("Unexpected config value: %+v, want: %+v", cfg, expectedCfg)
	//}
}
