package main

import (
	"github.com/gin-gonic/gin"
	"supernova/pkg/constance"
	"supernova/pkg/discovery"
	"supernova/scheduler/util"
)

type Executor struct {
	discoveryClient discovery.Client
}

func (e *Executor) Start() error {
	ins := new(discovery.ServiceInstance)
	ins.Host = GrpcHost
	ins.Port = GrpcPort
	ins.Meta = make(map[string]string, 3)
	ins.Meta[constance.ExecutorTagFieldName] = util.EncodeTag([]string{"tag-shell", "tag-1"})
	ins.Protoc = discovery.ProtocTypeGrpc
	ins.ServiceName = constance.ExecutorServiceName
	ins.MiddlewareHealthCheckUrl = "9.134.5.191:10001"

	router := gin.Default()

	// 添加一个用于健康检查的路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// 监听指定的 IP 地址和端口
	go func() {
		err := router.Run(":10001")
		if err != nil {
			panic(err)
		}
	}()

	return e.discoveryClient.Register(ins)
}
