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
