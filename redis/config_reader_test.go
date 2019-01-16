package main

import (
    "testing"
	"io/ioutil"
	"os"
	"reflect"
)

func TestConfigReader_FromFile(t *testing.T) {
    tests := []struct {
        name    string
        config  string
        expectedCfg *RedisConfig
    } {
        {
            name: "single-redis-test-server",
            config: `127.0.0.1:6379`,
            expectedCfg: &RedisConfig {
                servers: []*RedisConfigServer{
                    &RedisConfigServer{
                        host: "127.0.0.1",
                        port: 6379,
                    },
                },
            },
        },
    }

    t.Parallel()

    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            tmpFile, err := ioutil.TempFile("/tmp", "single-redis-server-test-config")
            if err != nil {
                t.Fatalf("Failed to create a temp file: %s", err)
            }
            defer os.Remove(tmpFile.Name())

			if err := ioutil.WriteFile(tmpFile.Name(), []byte(testCase.config), 0644); err != nil {
				t.Fatalf("Failed to write the data to tmp file: %s", err)
			}

            cfg, err := ConfigFromFile(tmpFile.Name())
            if err != nil {
                t.Fatalf("Failed to read from the config: %s", err)
            }
            if !reflect.DeepEqual(cfg.servers, testCase.expectedCfg.servers) {
                t.Errorf("Diverging redis servers: want: %+v, got: %+v", cfg.servers, testCase.expectedCfg.servers)
            }
        })
    }
}
