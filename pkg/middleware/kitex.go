package middleware

import (
	"context"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
)

func PrintKitexRequestResponse(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request, response interface{}) error {
		klog.CtxTracef(ctx, "receive request: %+v\n", request)
		err := next(ctx, request, response)
		klog.CtxTracef(ctx, "response: %+v\n", request)
		return err
	}
}
