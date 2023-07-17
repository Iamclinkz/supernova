package main

import (
	api "supernova/pkg/api/executor"
	"supernova/pkg/conf"
	"supernova/pkg/discovery"
)

func main() {
	cfg := conf.GetCommonConfig(conf.Dev)

	executor := new(Executor)

	var err error
	if err = executor.Start(); err != nil {
		panic(err)
	}

	if executor.discoveryClient, err = discovery.NewDiscoveryClient(discovery.TypeConsul, &discovery.Config{
		Host: cfg.ConsulConf.Host,
		Port: cfg.ConsulConf.Port,
	}); err != nil {
		panic(err)
	}

	svr := api.NewServer(executor)

	err = svr.Run()

	if err != nil {
		panic(err)
	}
}
