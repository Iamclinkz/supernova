package trace

import (
	"supernova/pkg/conf"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
)

func InitProvider(serviceName string, instrumentConf *conf.OTelConf) provider.OtelProvider {
	ret := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint(instrumentConf.CollectorEndpointHost+":"+instrumentConf.CollectorEndpointPort),
		provider.WithInsecure(),
	)
	klog.Infof("init trace success with exporter address: %s, name:%s", instrumentConf.CollectorEndpointHost+":"+
		instrumentConf.CollectorEndpointPort, serviceName)
	return ret
}
