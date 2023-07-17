package trace

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"supernova/pkg/conf"
)

func InitProvider(serviceName string, instrumentConf *conf.OTelConf) provider.OtelProvider {
	ret := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint(instrumentConf.ExportEndpointHost+":"+instrumentConf.ExportEndpointPort),
		provider.WithInsecure(),
	)
	klog.Infof("init trace success with exporter address: %s, name:%s", instrumentConf.ExportEndpointHost+":"+
		instrumentConf.ExportEndpointPort, serviceName)
	return ret
}
