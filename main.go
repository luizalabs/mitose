package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/luizalabs/mitose/config"
	"github.com/luizalabs/mitose/controller"
	"github.com/luizalabs/mitose/gauge"
	"github.com/luizalabs/mitose/k8s"
)

func main() {
	currentNS, err := k8s.GetCurrentNamespace()
	if err != nil {
		printErrorAndExit("getting current namespace name", err)
	}
	configWatcher := getConfigWatcher(currentNS)

	go gauge.Run()
	for {
		ctx, cancel := context.WithCancel(context.Background())
		errChan := make(chan error)
		go func() { errChan <- run(ctx, currentNS) }()

		select {
		case err, ok := <-configWatcher:
			if err != nil {
				printErrorAndExit("watching configmap", err)
			}
			if !ok {
				configWatcher = getConfigWatcher(currentNS)
			}
			log.Println("rebuilding controllers")
			cancel()
		case err := <-errChan:
			log.Println("error received on errChan:", err)
			if err != nil && err != context.Canceled {
				printErrorAndExit("running controllers", err)
			}
		}
	}
}

func getConfigWatcher(namespace string) <-chan error {
	configWatcher, err := k8s.WatchConfigMap(namespace)
	if err != nil {
		printErrorAndExit("watching configmaps", err)
	}
	<-configWatcher // expected at least one config map
	return configWatcher
}

func run(ctx context.Context, currentNS string) error {
	configData, err := k8s.GetConfigMapData(currentNS, "config")
	if err != nil {
		return err
	}

	controllers := make([]*controller.Controller, 0)
	for _, v := range configData {
		conf := new(config.Config)
		if err := json.Unmarshal([]byte(v), conf); err != nil {
			return err
		}
		if !conf.Active {
			continue
		}
		c, err := controller.Factory(conf.Type, v)
		if err != nil {
			return err
		}
		controllers = append(controllers, c)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, currentController := range controllers {
		c := currentController
		g.Go(func() error { return c.Run(ctx) })
	}

	return g.Wait()
}

func printErrorAndExit(phase string, err error) {
	fmt.Fprintf(os.Stderr, "error %s: %s", phase, err)
	os.Exit(2)
}
