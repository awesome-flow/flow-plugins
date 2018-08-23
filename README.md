# [WIP] Flow Plugins

![logo](https://github.com/whiteboxio/flow/blob/master/flow.png)

This repository contains plugins for [Flow framework](https://github.com/whiteboxio/flow).

## Flow Plugin Infrastructure

Flow is a widely extendable software due to the plugin system. We use
Golang plugins in order to let developers create their custom links. A plugin
must conform to the same interface as the core links: expose a constructor that
will produce a new instance of the link. One is allowed to implement any kind
of link: receivers, links, sinks, multiplexers, demultiplexers, etc.

In order to link a plugin, one describes the component in the config file like:

```yaml
components:
  <link_name>:
    plugin: <plugin_name>
    constructor: <ConstrFunc>
    params:
      ...
```

`link_name` is the same concept as naming the core links: same plugin might be
instantiated as many times as needed under distinct names.

`plugin_name` (provided with no angle braces) is the name of the plugin. The
name of the plugin includes naming convention: it would be mapped to the real
file lookup path.

By default, plugins are expected to be found in folder called
`/etc/flowd/plugins`, but is configurable by specifying `FLOW_PLUGIN_PATH`
environment variable.

A structure of a plugin folder looks like:

```
/etc/flowd/plugins
└── plugin_name
    ├── plugin_name.go
    ├── plugin_name.so
    └── plugin_name_test.go
```

The .go files are pretty trivial with some minor remarks we will provide a bit
later.

The .so file is being created by go build once run with `-buildmode=plugin`.
For more details see [Golang plugin reference](https://golang.org/pkg/plugin/).

A plugin must be built for the same archetecture and with the same release of
Go. Frankly speaking, Go plugin ecosystem is pretty fragile on Darwin
architecture yet (the progres is quite promising as there is a great interest
in the community). Also, building your program with `GODEBUG=cgocheck=2` will
crash once you import plugin module (it drives go checkers crazy due to passing
Go pointers by non-Go runtime of shared object libraries). This is why we
strongly encourage developing and testing plugins on AMD64 architecture.

Below there is an example of a plugin

```go
package main

import (
	flow "github.com/whitebox/flow/pkg/core"
	"bufio"
	"fmt"
	"os"
)

type Stdout struct {
	Name   string
	buffer *bufio.Writer
	*flow.Connector
}

func NewStdout(name string, params flow.Params) (flow.Link, error) {
	writer := bufio.NewWriter(os.Stdout)
	return &Stdout{name, writer, flow.NewConnector()}, nil
}

func (s *Stdout) Recv(msg *flow.Message) error {
	s.buffer.Write([]byte(fmt.Sprintf("Message:\n"+
		"    meta: %+v\n"+
		"    payload: %s\n", msg.Meta, msg.Payload)))
	if flushErr := s.buffer.Flush(); flushErr != nil {
		return msg.AckFailed()
	}
	return msg.AckDone()
}

func main() {}
func init() {}
```

The major difference with regular links defined by flow core is:
  * package name is main
  * function main() is there to satisfy Go requirements
  * function init() is there to perform a static bootstrap (called once on
    plugin load)

## Copyright

This software is created by Oleg Sidorov in 2018.

This software is distributed under under MIT license. See LICENSE file for full license text.