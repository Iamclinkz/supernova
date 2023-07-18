package middleware

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/gin-gonic/gin"
)

func PrintGinHeader(c *gin.Context) {
	headers := c.Request.Header
	for key, values := range headers {
		for _, value := range values {
			klog.Tracef("%s: %s\n", key, value)
		}
	}
	c.Next()
}
