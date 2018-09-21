package main

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestConfigReader_ReadFile(t *testing.T) {

	configData := []byte(`
cluster test-cluster
	jump_fnv1a_ch replication 1
		flow_eventlog_tcp_7220:7220=000
		flow_eventlog_tcp_7221:7221=001
		flow_eventlog_tcp_7222:7222=002
		flow_eventlog_tcp_7223:7223=003
		flow_eventlog_tcp_7224:7224=004
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

	cfg, err := ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read the config: %s", err)
	}

	expectedCfg := &GraphiteConfig{
		clusters: []GraphiteConfigCluster{
			{},
		},
		routes: []GraphiteConfigRoute{
			{},
		},
	}
	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Fatalf("Unexpected config value: %+v, want: %+v", cfg, expectedCfg)
	}
}
