# Go - Rafting

[![Go Report Card](https://goreportcard.com/badge/github.com/danielgatis/go-rafting?style=flat-square)](https://goreportcard.com/report/github.com/danielgatis/go-rafting)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/danielgatis/go-rafting/master/LICENSE)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/danielgatis/go-rafting)

A framework to build clusters using the hashicorp's raft implementation.

<p align="center">
    <img src="example/example.gif" />
</p>

## Install

```bash
go get -u github.com/danielgatis/go-rafting
```

And then import the package in your code:

```go
import "github.com/danielgatis/go-rafting"
```

### Usage

Create a new app state with some commands:

```go
state := rafting.NewState(make(map[string]string))

state.Command("set", func(data interface{}, args map[string]interface{}) (interface{}, error) {
    d := data.(map[string]string)
    key := cast.ToString(args["key"])
    val := cast.ToString(args["val"])
    d[key] = val
    return val, nil
})

state.Command("get", func(data interface{}, args map[string]interface{}) (interface{}, error) {
    d := data.(map[string]string)
    key := cast.ToString(args["key"])
    return d[key], nil
})

state.Command("del", func(data interface{}, args map[string]interface{}) (interface{}, error) {
    d := data.(map[string]string)
    key := cast.ToString(args["key"])
    value := d[key]
    delete(d, key)
    return value, nil
})
```

Create a new raft node with a mdns discovery:

```go
node, err := rafting.NewNode(rid, state.FSM(), rport, rafting.WithMdnsDiscovery())
if err != nil {
    logrus.Fatal(err)
}
```

Starting the node:

```go
node.Start(context.Background())
```

Thats it!

### Example

The example bellow is the code for the banner video.

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/danielgatis/go-ctrlc"
	"github.com/danielgatis/go-rafting"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

var (
	rid   string
	rport int
	hport int
)

func init() {
	flag.StringVar(&rid, "rid", "1", "raft node id")
	flag.IntVar(&rport, "rport", 4001, "raft port number")
	flag.IntVar(&hport, "hport", 3001, "http port number")
}

func main() {
	flag.Parse()

	// raft node
	state := rafting.NewState(make(map[string]string))
	node, err := rafting.NewNode(rid, state.FSM(), rport, rafting.WithMdnsDiscovery())
	if err != nil {
		logrus.Fatal(err)
	}

	state.Command("set", func(data interface{}, args map[string]interface{}) (interface{}, error) {
		d := data.(map[string]string)
		key := cast.ToString(args["key"])
		val := cast.ToString(args["val"])
		d[key] = val
		return val, nil
	})

	state.Command("get", func(data interface{}, args map[string]interface{}) (interface{}, error) {
		d := data.(map[string]string)
		key := cast.ToString(args["key"])
		return d[key], nil
	})

	state.Command("del", func(data interface{}, args map[string]interface{}) (interface{}, error) {
		d := data.(map[string]string)
		key := cast.ToString(args["key"])
		value := d[key]
		delete(d, key)
		return value, nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() { logrus.Fatal(node.Start(ctx)) }()

	// http server
	app := fiber.New()
	app.Get("/:op/:key/:val?", func(c *fiber.Ctx) error {
		result, err := node.Apply(c.Params("op"), time.Second, "key", c.Params("key"), "val", c.Params("val"))
		if err != nil {
			return err
		}

		return c.SendString(cast.ToString(result))
	})

	go func() { logrus.Fatal(app.Listen(fmt.Sprintf(":%d", hport))) }()

	// waiting
	ctrlc.Watch(func() {
		cancel()
		app.Shutdown()
	})

	<-ctx.Done()
}
```

### License

Copyright (c) 2021-present [Daniel Gatis](https://github.com/danielgatis)

Licensed under [MIT License](./LICENSE)

### Buy me a coffee

Liked some of my work? Buy me a coffee (or more likely a beer)

<a href="https://www.buymeacoffee.com/danielgatis" target="_blank"><img src="https://bmc-cdn.nyc3.digitaloceanspaces.com/BMC-button-images/custom_images/orange_img.png" alt="Buy Me A Coffee" style="height: auto !important;width: auto !important;"></a>
