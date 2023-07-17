package conf

var DevTraceConfig *OTelConf = &OTelConf{
	ExportEndpointHost: "localhost",
	ExportEndpointPort: "4317",
}

var K8sTraceConfig *OTelConf = &OTelConf{
	ExportEndpointHost: "9.134.5.191",
	ExportEndpointPort: "4317",
}

type OTelConf struct {
	ExportEndpointHost string
	ExportEndpointPort string
}
