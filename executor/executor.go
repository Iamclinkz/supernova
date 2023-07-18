package main

import (
	"supernova/pkg/constance"
	"supernova/pkg/discovery"

	"github.com/gin-gonic/gin"
)

type Executor struct {
	discoveryClient discovery.Client
}

func (e *Executor) Start(tags string, healthPort string, grpcPort int) error {
	ins := new(discovery.ServiceInstance)
	ins.Host = GrpcHost
	ins.Port = grpcPort
	ins.Meta = make(map[string]string, 1)
	//ins.Meta[constance.ExecutorTagFieldName] = util.EncodeTag([]string{"tagShell", "tagA"})
	ins.Meta[constance.ExecutorTagFieldName] = tags
	ins.Protoc = discovery.ProtocTypeGrpc
	ins.ServiceName = constance.ExecutorServiceName
	ins.InstanceId = "Executor-" + healthPort
	ins.MiddlewareHealthCheckUrl = "http://9.134.5.191:" + healthPort + "/health"

	router := gin.Default()

	// 添加一个用于健康检查的路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	// 监听指定的 IP 地址和端口
	go func() {
		err := router.Run(":" + healthPort)
		if err != nil {
			panic(err)
		}
	}()

	return e.discoveryClient.Register(ins)
}
