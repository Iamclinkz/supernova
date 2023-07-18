package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	api "supernova/pkg/api/executor"
	"supernova/pkg/conf"
	"supernova/pkg/discovery"
	"supernova/pkg/middleware"

	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func main() {
	cfg := conf.GetCommonConfig(conf.Dev)

	executor := new(Executor)

	var err error
	if executor.discoveryClient, err = discovery.NewDiscoveryClient(discovery.TypeConsul, &discovery.Config{
		Host: cfg.ConsulConf.Host,
		Port: cfg.ConsulConf.Port,
	}); err != nil {
		panic(err)
	}

	var tag string
	var healthPort string
	var grpcPort int

	flag.StringVar(&tag, "tag", "", "")
	flag.StringVar(&healthPort, "healthPort", "", "")
	flag.IntVar(&grpcPort, "grpcPort", 0, "")

	flag.Parse()

	fmt.Println(tag)

	if err = executor.Start(tag, healthPort, grpcPort); err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(grpcPort))
	if err != nil {
		panic(err)
	}

	svr := api.NewServer(executor,
		server.WithServiceAddr(addr),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithMiddleware(middleware.PrintKitexRequestResponse),
	)

	err = svr.Run()

	if err != nil {
		panic(err)
	}
}
