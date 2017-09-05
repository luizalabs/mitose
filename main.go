package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/controller"
	"github.com/luizalabs/mitose/gauge"
	"github.com/luizalabs/mitose/k8s"
)

func main() {
	awsKey := os.Getenv("AWS_KEY")
	awsSecret := os.Getenv("AWS_SECRET")
	awsRegion := os.Getenv("AWS_REGION")
	defaultInterval := os.Getenv("INTERVAL")

	configData, err := k8s.GetConfigMapData("mitose", "config")
	if err != nil {
		printErrorAndExit("getting config from config maps", err)
	}

	controllers := make([]*controller.Controller, 0)
	for _, v := range configData {
		conf := new(config.Config)
		if err := json.Unmarshal([]byte(v), conf); err != nil {
			printErrorAndExit("unmarshing json", err)
		}
		switch conf.Type {
		case "sqs":
			c, err := controller.NewSQSController(awsKey, awsSecret, awsRegion, v)
			if err != nil {
				printErrorAndExit("creating controller", err)
			}
			controllers = append(controllers, c)
		}
	}

	interval, err := time.ParseDuration(defaultInterval)
	if err != nil {
		printErrorAndExit("parsing time", err)
	}

	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)
	for _, currentController := range controllers {
		c := currentController
		g.Go(func() error { return c.Run(ctx, interval) })
	}
	g.Go(gauge.Run)

	if err = g.Wait(); err != nil {
		printErrorAndExit("running controllers", err)
	}
}

func printErrorAndExit(phase string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %s", phase, err)
	os.Exit(2)
}
